-- name: CreateLog :one
INSERT INTO logs (service_name, log_level, message)
VALUES ($1, $2, $3)
RETURNING *;
