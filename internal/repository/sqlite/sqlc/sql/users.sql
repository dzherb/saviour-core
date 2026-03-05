-- name: CreateUser :one
INSERT INTO users (uuid, username, password_hash)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE uuid = ?;