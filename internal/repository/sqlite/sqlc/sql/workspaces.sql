-- name: CreateWorkspace :one
INSERT INTO workspaces (
  uuid,
  name,
  author_uuid
)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET name = ?
WHERE uuid = ?
RETURNING *;

-- name: AddUserToWorkspace :exec
INSERT INTO workspace_users (
  user_uuid,
  workspace_uuid
)
VALUES (?, ?);

-- name: RemoveUserFromWorkspace :execrows
DELETE FROM workspace_users
WHERE user_uuid = ?
  AND workspace_uuid = ?;

-- name: DoesUserBelongToWorkspace :one
SELECT CAST(
  EXISTS(
    SELECT 1
    FROM workspace_users
    WHERE user_uuid = ? AND workspace_uuid = ?
  ) AS BOOLEAN
);
