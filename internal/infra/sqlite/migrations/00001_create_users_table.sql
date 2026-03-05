-- +goose Up
CREATE TABLE users (
    uuid UUID_BLOB PRIMARY KEY,
    created_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),
    updated_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),

    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT 0 CHECK (is_admin IN (0,1))
) WITHOUT ROWID;

-- +goose StatementBegin
CREATE TRIGGER users_set_updated_at_trigger
    AFTER UPDATE ON users
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE users
    SET updated_at = unixepoch()
    WHERE uuid = OLD.uuid;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER users_set_updated_at_trigger;

DROP TABLE users;
