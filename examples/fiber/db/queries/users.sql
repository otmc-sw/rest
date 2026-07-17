-- name: CreateUser :one
INSERT INTO users (
    username, 
    full_name, 
    email, 
    enabled, 
    test_int, 
    content, 
    test_string_array, 
    test_int_array, 
    test_map, 
    test_json, 
    profile
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUser :one
SELECT 
id, 
username, 
full_name, 
email, 
enabled, 
content, 
test_int, 
test_string_array, 
test_int_array, 
test_map, 
test_json, 
profile, 
created_at, 
updated_at 
FROM users WHERE id = ?;

-- name: GetAllUsers :many
SELECT 
id, 
username, 
full_name, 
email, 
enabled, 
content, 
test_int, 
test_string_array, 
test_int_array, 
test_map, 
test_json, 
profile, 
created_at, 
updated_at 
FROM users;

-- name: UpdateUser :one
UPDATE users
SET
    username = COALESCE(sqlc.narg(username), username),
    full_name = COALESCE(sqlc.narg(full_name), full_name),
    email = COALESCE(sqlc.narg(email), email),
    content = COALESCE(sqlc.narg(content), content),
    test_int = COALESCE(sqlc.narg(test_int), test_int),
    test_string_array = COALESCE(sqlc.narg(test_string_array), test_string_array),
    test_int_array = COALESCE(sqlc.narg(test_int_array), test_int_array),
    test_map = COALESCE(sqlc.narg(test_map), test_map),
    test_json = COALESCE(sqlc.narg(test_json), test_json),
    profile = COALESCE(sqlc.narg(profile), profile),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT 
id, 
username, 
full_name, 
email, 
enabled, 
content, 
test_int, 
test_string_array, 
test_int_array, 
test_map, 
test_json, 
profile, 
created_at, 
updated_at 
FROM users;
