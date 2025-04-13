-- name: GetUserByLogin :one
SELECT id, login, password, created_at, updated_at FROM users WHERE login=$1 LIMIT 1;

-- name: GetUserById :one
SELECT id, login, password, created_at, updated_at FROM users WHERE id=$1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (login, password, created_at, updated_at)
VALUES ($1, $2, $3, $3)
RETURNING id;
