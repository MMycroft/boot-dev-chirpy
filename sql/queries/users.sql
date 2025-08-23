-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password, is_chirpy_red)
VALUES (GEN_RANDOM_UUID(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $1, $2, $3)
RETURNING *;

-- name: GetUsers :many
SELECT *
FROM users
ORDER BY created_at ASC;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE users.id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE users.email = $1;

-- name: UpdateUser :one
UPDATE users
SET 
  updated_at = CURRENT_TIMESTAMP,
  email = $2,
  hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: UpgradeUser :one
UPDATE users
SET 
  updated_at = CURRENT_TIMESTAMP,
  is_chirpy_red = $1
WHERE id = $1
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: DeleteUser :exec
DELETE FROM users
WHERE users.id = $1;
