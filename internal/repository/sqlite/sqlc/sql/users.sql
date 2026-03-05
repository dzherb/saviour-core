-- name: CreateUser :one
INSERT INTO users (uuid, username, password_hash)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetUserByUUID :one
SELECT * FROM users WHERE uuid = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;
