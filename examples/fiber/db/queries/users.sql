-- name: CreateUser :exec
INSERT INTO users (username, full_name, email) VALUES (?, ?, ?);

-- name: GetUser :one
SELECT id, username, full_name, email, created_at, updated_at FROM users WHERE id = ?;

-- name: GetAllUsers :many
SELECT id, username, full_name, email, created_at, updated_at FROM users;

-- name: UpdateUser :exec
UPDATE users SET username = ?, full_name = ?, email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT id, username, full_name, email, created_at, updated_at FROM users;