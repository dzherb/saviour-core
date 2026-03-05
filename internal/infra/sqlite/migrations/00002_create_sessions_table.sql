-- +goose Up
CREATE TABLE sessions
(
    uuid UUID_BLOB PRIMARY KEY,
    created_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),
    updated_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),

    user_uuid UUID_BLOB NOT NULL,
    refresh_token_hash TEXT NOT NULL,
    revoked_at UNIXTIME_INT NULL,
    last_refresh_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),
    user_agent TEXT NOT NULL,
    ip_last IP_BLOB NOT NULL,

    FOREIGN KEY (user_uuid) REFERENCES users(uuid) ON DELETE CASCADE
) WITHOUT ROWID;

-- +goose StatementBegin
CREATE TRIGGER auth_sessions_set_updated_at_trigger
    AFTER UPDATE ON sessions
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE sessions
    SET updated_at = unixepoch()
    WHERE uuid = OLD.uuid;
END;
-- +goose StatementEnd

CREATE INDEX auth_sessions_user_uuid_fk ON sessions(user_uuid);

-- +goose Down
DROP INDEX auth_sessions_user_uuid_fk;

DROP TRIGGER auth_sessions_set_updated_at_trigger;

DROP TABLE sessions;