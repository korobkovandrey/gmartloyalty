package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/korobkovandrey/gmartloyalty/db/query"
	"github.com/korobkovandrey/gmartloyalty/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuth_Login(t *testing.T) {
	ctx := context.Background()
	cfg := &AuthConfig{
		Secret:   "test-secret",
		LifeTime: 1 * time.Hour,
	}
	login := "testuser"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("successful login", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{ID: 1, Login: login, Password: string(hashedPassword)}, nil)

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Login(ctx, login, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.True(t, len(token) > len("Bearer "))
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{}, sql.ErrNoRows)

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Login(ctx, login, password)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Empty(t, token)
	})

	t.Run("wrong password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{ID: 1, Login: login, Password: string(hashedPassword)}, nil)

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Login(ctx, login, "wrongpassword")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Empty(t, token)
	})

	t.Run("store error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storeErr := errors.New("database error")
		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{}, storeErr)

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Login(ctx, login, password)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), storeErr.Error())
		assert.Empty(t, token)
	})
}

func TestAuth_Register(t *testing.T) {
	ctx := context.Background()
	cfg := &AuthConfig{
		Secret:   "test-secret",
		LifeTime: 1 * time.Hour,
	}
	login := "testuser"
	password := "password123"

	t.Run("successful registration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{}, sql.ErrNoRows)
		mockStore.EXPECT().
			CreateUser(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, params query.CreateUserParams) (int64, error) {
				assert.Equal(t, login, params.Login)
				assert.NotEmpty(t, params.Password)
				assert.WithinDuration(t, time.Now().UTC(), params.CreatedAt, 1*time.Second)
				return 1, nil
			})

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Register(ctx, login, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.True(t, len(token) > len("Bearer "))
	})

	t.Run("login already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{ID: 1, Login: login}, nil)

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Register(ctx, login, password)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrAlreadyExists))
		assert.Empty(t, token)
	})

	t.Run("integrity constraint violation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{}, sql.ErrNoRows)
		mockStore.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(int64(0), &pgconn.PgError{Code: pgerrcode.UniqueViolation})

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Register(ctx, login, password)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrAlreadyExists))
		assert.Empty(t, token)
	})

	t.Run("store error on check login", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storeErr := errors.New("database error")
		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{}, storeErr)

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Register(ctx, login, password)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), storeErr.Error())
		assert.Empty(t, token)
	})

	t.Run("store error on create user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storeErr := errors.New("database error")
		mockStore := mocks.NewMockUserAuthStore(ctrl)
		mockStore.EXPECT().
			GetUserByLogin(ctx, login).
			Return(query.User{}, sql.ErrNoRows)
		mockStore.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(int64(0), storeErr)

		auth := NewAuth(cfg, mockStore)
		token, err := auth.Register(ctx, login, password)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), storeErr.Error())
		assert.Empty(t, token)
	})
}

func TestAuth_sign(t *testing.T) {
	cfg := &AuthConfig{
		Secret:   "test-secret",
		LifeTime: 1 * time.Hour,
	}
	auth := NewAuth(cfg, nil)

	t.Run("successful signing", func(t *testing.T) {
		token, err := auth.sign(1)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.True(t, len(token) > len("Bearer "))

		parsedToken, err := jwt.Parse(token[len("Bearer "):], func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})
		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, float64(1), claims["userID"])
		assert.NotEmpty(t, claims["exp"])
	})

	t.Run("no lifetime", func(t *testing.T) {
		authNoLifetime := NewAuth(&AuthConfig{Secret: "test-secret"}, nil)
		token, err := authNoLifetime.sign(1)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		parsedToken, err := jwt.Parse(token[len("Bearer "):], func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, float64(1), claims["userID"])
		assert.Nil(t, claims["exp"])
	})
}

func Test_makePasswordHash(t *testing.T) {
	t.Run("successful hash", func(t *testing.T) {
		password := "password123"
		hash, err := makePasswordHash(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		assert.NoError(t, err)
	})

	t.Run("empty password", func(t *testing.T) {
		hash, err := makePasswordHash("")
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(""))
		assert.NoError(t, err)
	})

	t.Run("long password", func(t *testing.T) {
		_, err := makePasswordHash(strings.Repeat("a", 100))
		require.Error(t, err)
	})
}
