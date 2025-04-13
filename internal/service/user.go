package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/korobkovandrey/gmartloyalty/db/query"
)

type UserFinderStore interface {
	GetUserById(context.Context, int64) (query.User, error)
}

type UserFinder struct {
	r UserFinderStore
}

func NewUserFinder(r UserFinderStore) *UserFinder {
	return &UserFinder{r: r}
}

func (s *UserFinder) Find(ctx context.Context, id int64) (*query.User, error) {
	u, err := s.r.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get user: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}
