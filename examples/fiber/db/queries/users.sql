-- name: CreateUser :one
INSERT INTO users (username, full_name, email, enabled, test_int, content) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUser :one
SELECT id, username, full_name, email, enabled, content, test_int, created_at, updated_at FROM users WHERE id = ?;

-- name: GetAllUsers :many
SELECT id, username, full_name, email, enabled, content, test_int, created_at, updated_at FROM users;

-- name: UpdateUser :one
UPDATE users
SET
    username = COALESCE(sqlc.narg(username), username),
    full_name = COALESCE(sqlc.narg(full_name), full_name),
    email = COALESCE(sqlc.narg(email), email),
    content = COALESCE(sqlc.narg(content), content),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT id, username, full_name, email, enabled, content, test_int, created_at, updated_at FROM users;