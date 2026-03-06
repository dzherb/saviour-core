-- +goose Up
CREATE TABLE workspace_users
(
    created_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),
    updated_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),

    user_uuid UUID_BLOB NOT NULL,
    workspace_uuid UUID_BLOB NOT NULL,

    PRIMARY KEY (user_uuid, workspace_uuid),

    FOREIGN KEY (workspace_uuid) REFERENCES workspaces(uuid) ON DELETE CASCADE,
    FOREIGN KEY (user_uuid) REFERENCES users(uuid) ON DELETE CASCADE
) WITHOUT ROWID;

-- +goose StatementBegin
CREATE TRIGGER workspace_users_set_updated_at_trigger
    AFTER UPDATE ON workspace_users
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE workspace_users
    SET updated_at = unixepoch()
    WHERE uuid = OLD.uuid;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER workspace_users_set_updated_at_trigger;

DROP TABLE workspace_users;