-- name: Accrual :one
SELECT COALESCE(SUM(accrual), 0)::numeric FROM orders WHERE status = 'PROCESSED' AND user_id = $1;

-- name: Withdrawn :one
SELECT COALESCE(SUM(sum), 0)::numeric FROM withdrawals WHERE user_id = $1;
