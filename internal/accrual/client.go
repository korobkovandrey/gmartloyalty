package accrual

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"resty.dev/v3"
)

type client struct {
	*resty.Client
}

func newClient(baseURL string) *client {
	return &client{resty.New().
		SetBaseURL(baseURL).
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)}
}

type orderResponse struct {
	Order      string
	Status     string
	Accrual    float64
	RetryAfter int
}

func (c *client) getOrder(ctx context.Context, order string) (*orderResponse, error) {
	result := &orderResponse{}
	r, err := c.R().SetContext(ctx).
		SetResult(result).Get("/api/orders/" + order)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if r.StatusCode() == http.StatusTooManyRequests {
		result.RetryAfter = 60
		if ra := r.Header().Get("Retry-After"); ra != "" {
			if raInt, err := strconv.Atoi(ra); err == nil {
				result.RetryAfter = raInt
			}
		}
	}
	return result, nil
}

func (c *client) close() {
	_ = c.Client.Close()
}
