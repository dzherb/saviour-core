-- +goose Up
CREATE TABLE workspaces
(
    uuid UUID_BLOB PRIMARY KEY,
    created_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),
    updated_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),

    name TEXT NOT NULL CHECK ( length(name) != 0 ),
    author_uuid UUID_BLOB NOT NULL,

    FOREIGN KEY (author_uuid) REFERENCES users(uuid) ON DELETE RESTRICT
) WITHOUT ROWID;

-- +goose StatementBegin
CREATE TRIGGER workspaces_set_updated_at_trigger
    AFTER UPDATE ON workspaces
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE workspaces
    SET updated_at = unixepoch()
    WHERE uuid = OLD.uuid;
END;
-- +goose StatementEnd

CREATE INDEX workspaces_author_uuid_fk ON workspaces(author_uuid);

-- +goose Down
DROP INDEX workspaces_author_uuid_fk;

DROP TRIGGER workspaces_set_updated_at_trigger;

DROP TABLE workspaces;