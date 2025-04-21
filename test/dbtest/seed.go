package dbtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const (
	UserID1            = 1
	UserID2            = 2
	ProcessedOrderNum  = "1345821735824"
	ProcessedOrderNum1 = "5300850608"
	ProcessedOrderNum2 = "7214322575"
	NewOrderNum        = "8610513031"
	InvalidOrderNum    = "4720781618"
	ProcessingOrderNum = "36280085477"

	WithdrawOrderNum  = "3747485070"
	WithdrawOrderNum1 = "436048"
	WithdrawOrderNum2 = "6828420"
	WithdrawOrderNum3 = "60638327573147"
)

func SeedUsers(t *testing.T, db *sql.DB) {
	t.Helper()
	bytes, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	ctx := context.TODO()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users(id, login, password, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $4);`, UserID1, "user"+string(rune(UserID1)), string(bytes), time.Now())
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO users(id, login, password, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $4);`, UserID2, "user"+string(rune(UserID2)), string(bytes), time.Now())
	require.NoError(t, err)
}

func SeedOrders(t *testing.T, db *sql.DB) {
	t.Helper()
	type orderRow struct {
		userID  int64
		number  string
		status  string
		accrual float64
	}
	orders := []orderRow{
		{userID: UserID1, number: ProcessedOrderNum, status: "PROCESSED", accrual: 100.5},
		{userID: UserID1, number: ProcessedOrderNum1, status: "PROCESSED", accrual: 60.3},
		{userID: UserID1, number: NewOrderNum, status: "NEW", accrual: 0},
		{userID: UserID1, number: InvalidOrderNum, status: "INVALID", accrual: 0},
		{userID: UserID1, number: ProcessingOrderNum, status: "PROCESSING", accrual: 0},
		{userID: UserID2, number: ProcessedOrderNum2, status: "PROCESSED", accrual: 67.8},
	}
	ctx := context.TODO()
	for _, o := range orders {
		_, err := db.ExecContext(ctx, `
			INSERT INTO orders(user_id, number, status, accrual, uploaded_at)
			VALUES
				($1, $2, $3, $4, $5);`, o.userID, o.number, o.status, o.accrual, time.Now())
		require.NoError(t, err)
	}
}

func SeedWithdraw(t *testing.T, db *sql.DB) {
	type withdrawRow struct {
		userID int64
		order  string
		sum    float64
	}
	withdrawals := []withdrawRow{
		{userID: UserID1, order: WithdrawOrderNum, sum: 10.2},
		{userID: UserID1, order: WithdrawOrderNum1, sum: 20.3},
		{userID: UserID2, order: WithdrawOrderNum2, sum: 12.6},
		{userID: UserID2, order: WithdrawOrderNum3, sum: 5.4},
	}
	ctx := context.TODO()
	for _, w := range withdrawals {
		_, err := db.ExecContext(ctx, `
			INSERT INTO withdrawals(user_id, "order", sum, processed_at)
			VALUES
				($1, $2, $3, $4);`, w.userID, w.order, w.sum, time.Now())
		require.NoError(t, err)
	}
}
