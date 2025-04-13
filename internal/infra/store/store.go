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
	db *sql.DB
}

func NewStore(ctx context.Context, dsn string) (*Store, error) {
	if err := db.Migrate(dsn); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	s := &Store{}
	var err error
	s.db, err = sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	s.db.SetMaxOpenConns(5)
	s.db.SetMaxIdleConns(5)
	s.db.SetConnMaxIdleTime(time.Minute)
	s.db.SetConnMaxLifetime(time.Minute)
	err = s.db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	s.Queries = query.New(s.db)
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
