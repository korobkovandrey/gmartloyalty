package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/korobkovandrey/gmartloyalty/db"
	"github.com/korobkovandrey/gmartloyalty/db/query"
)

type Store struct {
	*query.Queries
	DB *sql.DB
}

func NewStore(ctx context.Context, dsn string) (*Store, error) {
	if err := db.Migrate(dsn); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	s := &Store{}
	var err error
	s.DB, err = sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	s.DB.SetMaxOpenConns(5)
	s.DB.SetMaxIdleConns(5)
	s.DB.SetConnMaxIdleTime(time.Minute)
	s.DB.SetConnMaxLifetime(time.Minute)
	err = s.DB.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	s.Queries = query.New(s.DB)
	return s, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func (s *Store) Balance(ctx context.Context, userID int64) (float64, error) {
	a, err := s.Accrual(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get accrual: %w", err)
	}
	w, err := s.Withdrawn(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get withdrawn: %w", err)
	}
	return a - w, nil
}
