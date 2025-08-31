-- name: InsertFreshToken :one
INSERT INTO refreshtokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    NOW() + INTERVAL '60 days',
    NULL
)
RETURNING *;


-- name: ResetRefreshTokens :exec
DELETE FROM refreshtokens;


-- name: GetFreshTokenByToken :one
SELECT * FROM refreshtokens WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refreshtokens
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1;