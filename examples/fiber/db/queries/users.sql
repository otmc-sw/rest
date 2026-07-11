-- name: CreateUser :one
INSERT INTO users (username, full_name, email, content) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUser :one
SELECT id, username, full_name, email, content, created_at, updated_at FROM users WHERE id = ?;

-- name: GetAllUsers :many
SELECT id, username, full_name, email, content, created_at, updated_at FROM users;

-- name: UpdateUser :exec
UPDATE users SET username = ?, full_name = ?, email = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT id, username, full_name, email, content, created_at, updated_at FROM users;