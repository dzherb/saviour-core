package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/netip"
	"time"

	"github.com/google/uuid"

	"saviour/internal/model"
)

// DBTX represents *sql.DB.
type DBTX interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type UserRepository interface {
	CreateUser(
		ctx context.Context,
		params CreateUserParams,
	) (*model.User, error)
	GetUserByUsername(
		ctx context.Context,
		username string,
	) (*model.User, error)
}

var (
	ErrUserNotFound = errors.New("user not found")
)

type CreateUserParams struct {
	UUID         uuid.UUID
	Username     string
	PasswordHash string
}

type GetUserByCredentialsParams struct {
	Username     string
	PasswordHash string
}

type SessionRepository interface {
	CreateSession(
		ctx context.Context,
		params CreateSessionParams,
	) (*model.Session, error)
	RefreshActiveSession(
		ctx context.Context,
		params RefreshActiveSessionParams,
	) error
	RevokeSession(
		ctx context.Context,
		params RevokeSessionParams,
	) error
}

var (
	ErrSessionNotFound = errors.New("session not found")
)

type CreateSessionParams struct {
	UUID             uuid.UUID
	UserUUID         uuid.UUID
	RefreshTokenHash string
	UserAgent        string
	IP               netip.Addr
}

type RefreshActiveSessionParams struct {
	UUID                uuid.UUID
	RefreshTokenHash    string
	NewRefreshTokenHash string
	IP                  netip.Addr
	RefreshTTL          time.Duration
}

type RevokeSessionParams struct {
	UUID     uuid.UUID
	UserUUID uuid.UUID
}
