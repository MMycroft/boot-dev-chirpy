-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $2, CURRENT_TIMESTAMP + INTERVAL '60 day', NULL)
RETURNING *;

-- name: GetRefreshTokens :many
SELECT *
FROM refresh_tokens
WHERE CURRENT_TIMESTAMP < expires_at
  AND revoked_at IS NULL
ORDER BY created_at ASC;

-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1
  AND CURRENT_TIMESTAMP < expires_at
  AND revoked_at IS NULL;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET
  revoked_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE token = $1
RETURNING revoked_at;

-- name: DeleteRefreshTokens :exec
DELETE FROM refresh_tokens;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;
