-- name: CreateUser :exec
INSERT INTO users (username, email) VALUES (?, ?);

-- name: GetUser :one
SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?;

-- name: UpdateUser :exec
UPDATE users SET username = ?, email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT id, username, email, created_at, updated_at FROM users;