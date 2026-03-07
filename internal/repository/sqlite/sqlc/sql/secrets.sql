-- name: CreateSecret :one
INSERT INTO secrets (
  uuid,
  name,
  value_encrypted,
  author_uuid,
  workspace_uuid
)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: RenameSecret :one
UPDATE secrets
SET name = ?
WHERE uuid = ?
RETURNING *;

-- name: UpdateSecret :one
UPDATE secrets
SET name = ?,
  value_encrypted = ?
WHERE uuid = ?
RETURNING *;