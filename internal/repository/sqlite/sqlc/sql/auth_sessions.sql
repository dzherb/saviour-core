-- name: CreateSession :one
INSERT INTO sessions (
    uuid,
    user_uuid,
    refresh_token_hash,
    user_agent,
    ip_last
)
VALUES (?, ?, ?, ?, sqlc.arg('IP')) RETURNING *;

-- name: RefreshActiveSession :execrows
UPDATE sessions
SET refresh_token_hash = ?,
    last_refresh_at = unixepoch(),
    ip_last = sqlc.arg('IP')
WHERE uuid = ?
  AND revoked_at IS NULL
  -- session is not expired
  AND last_refresh_at > (unixepoch() - CAST(sqlc.arg('refreshTTLInSec') AS INTEGER));

-- name: RevokeSession :execrows
UPDATE sessions
SET revoked_at = unixepoch()
WHERE uuid = ?
  AND user_uuid = ?
  AND revoked_at IS NULL;