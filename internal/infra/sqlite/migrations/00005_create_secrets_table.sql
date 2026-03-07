-- +goose Up
CREATE TABLE secrets
(
    uuid UUID_BLOB PRIMARY KEY,
    created_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),
    updated_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),

    name TEXT NOT NULL CHECK ( length(name) != 0 ),
    value_encrypted BLOB NOT NULL,
    author_uuid UUID_BLOB NOT NULL,
    workspace_uuid UUID_BLOB NOT NULL,

    FOREIGN KEY (author_uuid) REFERENCES users(uuid) ON DELETE CASCADE,
    FOREIGN KEY (workspace_uuid) REFERENCES workspaces(uuid) ON DELETE CASCADE
) WITHOUT ROWID;

-- +goose StatementBegin
CREATE TRIGGER secrets_set_updated_at_trigger
    AFTER UPDATE ON secrets
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE secrets
    SET updated_at = unixepoch()
    WHERE uuid = OLD.uuid;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER secrets_set_updated_at_trigger;

DROP TABLE secrets;