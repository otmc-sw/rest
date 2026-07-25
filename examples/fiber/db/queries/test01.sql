-- name: CreateTest01 :one
INSERT INTO test01 
    (id, name, test_int) 
VALUES
    (?, ?, ?) 
RETURNING *;

-- name: GetTest01 :one
SELECT * FROM test01 WHERE id = ?;

-- name: GetAllTest01s :many
SELECT * FROM test01;

-- name: UpdateTest01 :one
UPDATE test01 SET
    name = COALESCE(sqlc.narg(name), name),
    test_int = COALESCE(sqlc.narg(test_int), test_int)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteTest01 :exec
DELETE FROM test01 WHERE id = ?;
