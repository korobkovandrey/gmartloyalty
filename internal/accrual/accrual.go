package accrual

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db/query"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
)

type orderStore interface {
	GetOrdersNotProcessed(ctx context.Context) ([]query.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (query.Order, error)
	SetOrderStatus(ctx context.Context, arg query.SetOrderStatusParams) error
	SetOrderStatusAndAccrual(ctx context.Context, arg query.SetOrderStatusAndAccrualParams) error
}

type job struct {
	orderNum string
	attempt  int
}

type Config struct {
	AccrualSystemAddress string
	JobsSize             int
	DeferJobsSize        int
	NumWorkers           int
	MaxAttempts          int
	AttemptTimeout       time.Duration
}

type Accrual struct {
	cfg         *Config
	jobsCh      chan job
	deferJobsCh chan job
	l           *logging.ZapLogger
	r           orderStore
	c           *client
	waitCh      chan struct{}
	waitMux     sync.RWMutex
}

func NewAccrual(l *logging.ZapLogger, r orderStore, cfg *Config) *Accrual {
	waitCh := make(chan struct{})
	close(waitCh)
	return &Accrual{
		cfg:         cfg,
		jobsCh:      make(chan job, cfg.JobsSize),
		deferJobsCh: make(chan job, cfg.DeferJobsSize),
		l:           l,
		r:           r,
		c:           newClient(cfg.AccrualSystemAddress),
		waitCh:      waitCh,
	}
}

func (s *Accrual) Run(ctx context.Context) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(s.cfg.NumWorkers + 1)
	go func() {
		for j := range s.deferJobsCh {
			if ctx.Err() != nil {
				break
			}
			s.pushJob(j)
		}
		wg.Done()
	}()
	for i := 0; i < s.cfg.NumWorkers; i++ {
		go func() {
			s.worker(ctx)
			wg.Done()
		}()
	}
	return wg
}

func (s *Accrual) PushOrder(orderNum string) {
	s.pushJob(job{orderNum: orderNum})
}

func (s *Accrual) pushJob(j job) {
	s.jobsCh <- j
}

func (s *Accrual) pushDeferredJob(j job) {
	time.AfterFunc(s.cfg.AttemptTimeout, func() {
		s.deferJobsCh <- j
	})
}

func (s *Accrual) getOrderToProcess(ctx context.Context, j job) (*query.Order, error) {
	order, err := s.r.GetOrderByNumber(ctx, j.orderNum)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			s.pushDeferredJob(j)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order.Status == query.TOrderStatusNEW {
		if err := s.r.SetOrderStatus(ctx, query.SetOrderStatusParams{
			ID:     order.ID,
			Status: query.TOrderStatusPROCESSING,
		}); err != nil {
			s.pushDeferredJob(j)
			return nil, fmt.Errorf("failed to set order status: %w", err)
		}
	} else if order.Status != query.TOrderStatusPROCESSING {
		return nil, errors.New("order is already processed")
	}
	return &order, nil
}

func (s *Accrual) worker(ctx context.Context) {
	for j := range s.jobsCh {
		s.waitRetry(ctx)
		if ctx.Err() != nil {
			break
		}
		order, err := s.getOrderToProcess(ctx, j)
		if err != nil {
			s.l.ErrorCtx(ctx, fmt.Errorf("failed to get order: %w", err).Error())
			continue
		}

		toCtx, toCancel := context.WithTimeout(ctx, 10*time.Second)
		resp, err := s.c.getOrder(toCtx, order.Number)
		toCancel()
		if err != nil {
			s.l.ErrorCtx(ctx, fmt.Errorf("failed to get accrual: %w", err).Error())
			resp = &orderResponse{
				Order: order.Number,
			}
		}

		if resp.Status == "" {
			j.attempt++
		}
		if resp.RetryAfter > 0 {
			s.startWaitRetry(ctx, resp.RetryAfter)
		}
		if resp.Status == "PROCESSED" || resp.Status == "INVALID" {
			if err := s.r.SetOrderStatusAndAccrual(ctx, query.SetOrderStatusAndAccrualParams{
				ID:      order.ID,
				Status:  query.TOrderStatus(resp.Status),
				Accrual: resp.Accrual,
			}); err != nil {
				s.l.ErrorCtx(ctx, fmt.Errorf("failed to set order status and accrual: %w", err).Error())
				s.pushDeferredJob(j)
				continue
			}
		} else if j.attempt < s.cfg.MaxAttempts {
			s.pushDeferredJob(j)
		} else {
			if err := s.r.SetOrderStatus(ctx, query.SetOrderStatusParams{
				ID:     order.ID,
				Status: query.TOrderStatusINVALID,
			}); err != nil {
				s.l.ErrorCtx(ctx, fmt.Errorf("failed to set order status: %w", err).Error())
				continue
			}
		}
	}
}

func (s *Accrual) waitRetry(ctx context.Context) {
	s.waitMux.RLock()
	defer s.waitMux.RUnlock()
	select {
	case <-ctx.Done():
	case <-s.waitCh:
	}
}

func (s *Accrual) startWaitRetry(ctx context.Context, timeout int) {
	s.waitMux.Lock()
	defer s.waitMux.Unlock()
	select {
	case <-s.waitCh:
		s.waitCh = make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
			case <-time.After(time.Duration(timeout) * time.Second):
			}
			close(s.waitCh)
		}()
	default:
	}
}

func (s *Accrual) Close() {
	close(s.deferJobsCh)
	close(s.jobsCh)
	s.c.close()
}
