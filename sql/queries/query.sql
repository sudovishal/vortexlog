-- name: CreateLog :one
INSERT INTO logs (service_name, log_level, message, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateLogs :copyfrom
INSERT INTO logs (service_name, log_level, message, created_at)
VALUES ($1, $2, $3,$4);