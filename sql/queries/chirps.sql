-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (GEN_RANDOM_UUID(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $1, $2)
RETURNING *;

-- name: GetChirps :many
SELECT *
FROM chirps
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT *
FROM chirps
WHERE chirps.id = $1;

-- name: DeleteChirps :exec
DELETE FROM chirps;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE chirps.id = $1;
