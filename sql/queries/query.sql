-- name: CreateLog :one
INSERT INTO logs (service_name, log_level, message, created_at, trace_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CreateLogs :copyfrom
INSERT INTO logs (service_name, log_level, message, created_at, trace_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6);