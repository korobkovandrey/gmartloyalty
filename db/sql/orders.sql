-- name: GetOrderByNumber :one
SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE number=$1 LIMIT 1;

-- name: GetOrderByID :one
SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE id=$1 LIMIT 1;

-- name: GetOrdersForUserID :many
SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE user_id=$1 ORDER BY uploaded_at DESC;

-- name: CreateOrder :one
INSERT INTO orders (user_id, number, status, uploaded_at)
VALUES ($1, $2,  'NEW', $3)
RETURNING id;

-- name: SetOrderStatus :exec
UPDATE orders SET status=$1 WHERE id=$2;

-- name: SetOrderStatusAndAccrual :exec
UPDATE orders SET status=$1, accrual=$2 WHERE id=$3;

-- name: GetOrdersNotProcessed :many
SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE status NOT IN ('PROCESSED', 'INVALID') ORDER BY uploaded_at;