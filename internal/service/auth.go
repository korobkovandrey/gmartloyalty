package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/korobkovandrey/gmartloyalty/db/query"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=auth.go -destination=mocks/mock_userauthstore.go -package=mocks

type UserAuthStore interface {
	GetUserByLogin(context.Context, string) (query.User, error)
	CreateUser(context.Context, query.CreateUserParams) (int64, error)
}

type AuthConfig struct {
	Secret   string
	LifeTime time.Duration
}

type Auth struct {
	r   UserAuthStore
	cfg *AuthConfig
}

func NewAuth(cfg *AuthConfig, r UserAuthStore) *Auth {
	return &Auth{r: r, cfg: cfg}
}

func (s *Auth) Login(ctx context.Context, login, password string) (string, error) {
	u, err := s.r.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("failed to get user: %w", ErrNotFound)
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("failed to verify password: %w: %w", ErrNotFound, err)
	}
	return s.sign(u.ID)
}

func (s *Auth) Register(ctx context.Context, login, password string) (string, error) {
	_, err := s.r.GetUserByLogin(ctx, login)
	if err == nil {
		return "", fmt.Errorf("login %w", ErrAlreadyExists)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("failed to check login: %w", err)
	}
	hash, err := makePasswordHash(password)
	if err != nil {
		return "", fmt.Errorf("failed to make password hash: %w", err)
	}
	id, err := s.r.CreateUser(ctx, query.CreateUserParams{
		Login:     login,
		Password:  hash,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && pgerrcode.IsIntegrityConstraintViolation(e.Code) {
			return "", fmt.Errorf("login %w", ErrAlreadyExists)
		}
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return s.sign(id)
}

func (s *Auth) sign(id int64) (string, error) {
	claims := jwt.MapClaims{
		"userID": id,
	}
	if s.cfg.LifeTime > 0 {
		claims["exp"] = time.Now().Add(s.cfg.LifeTime).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to create signed token: %w", err)
	}
	return "Bearer " + t, nil
}

func makePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), err
}
