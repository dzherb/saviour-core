-- name: CreateUser :one
INSERT INTO users (
  uuid,
  username,
  password_hash,
  is_admin
)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: DeactivateUser :execrows
UPDATE users
SET is_active = 0
WHERE uuid = ? AND is_active = 1;

-- name: GetUserByUUID :one
SELECT * FROM users WHERE uuid = ? AND is_active = 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ? AND is_active = 1;
