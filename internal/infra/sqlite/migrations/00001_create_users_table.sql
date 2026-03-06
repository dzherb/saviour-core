-- +goose Up
CREATE TABLE users (
    uuid UUID_BLOB PRIMARY KEY,
    created_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),
    updated_at UNIXTIME_INT NOT NULL DEFAULT (unixepoch()),

    username TEXT NOT NULL UNIQUE CHECK ( length(username) != 0 ),
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 1 CHECK (is_admin IN (0,1)),
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

CREATE INDEX users_username_idx ON users(username);

-- +goose Down
DROP INDEX users_username_idx;

DROP TRIGGER users_set_updated_at_trigger;

DROP TABLE users;
