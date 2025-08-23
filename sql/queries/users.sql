-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (GEN_RANDOM_UUID(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $1, $2)
RETURNING *;

-- name: GetUsers :many
SELECT *
FROM users
ORDER BY created_at ASC;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE users.id = $1;

-- name: UpdateUser :one
UPDATE users
SET 
  updated_at = CURRENT_TIMESTAMP,
  email = $2,
  hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE users.email = $1;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: DeleteUser :exec
DELETE FROM users
WHERE users.id = $1;
