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

var (
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

type UserRepository interface {
	CreateUser(
		ctx context.Context,
		params CreateUserParams,
	) (*model.User, error)
	DeactivateUser(
		ctx context.Context,
		uuid uuid.UUID,
	) error
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
	IsAdmin      bool
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

type WorkspaceRepository interface {
	CreateWorkspace(
		ctx context.Context,
		params CreateWorkspaceParams,
	) (*model.Workspace, error)
	UpdateWorkspace(
		ctx context.Context,
		params UpdateWorkspaceParams,
	) (*model.Workspace, error)
	AddUserToWorkspace(
		ctx context.Context,
		params AddUserToWorkspaceParams,
	) error
	RemoveUserFromWorkspace(
		ctx context.Context,
		params RemoveUserFromWorkspaceParams,
	) error
}

var (
	ErrWorkspaceNotFound     = errors.New("workspace not found")
	ErrWorkspaceUserNotFound = errors.New("workspace user not found")
)

type CreateWorkspaceParams struct {
	UUID       uuid.UUID
	Name       string
	AuthorUUID uuid.UUID
}

type UpdateWorkspaceParams struct {
	UUID uuid.UUID
	Name string
}

type AddUserToWorkspaceParams struct {
	UserUUID      uuid.UUID
	WorkspaceUUID uuid.UUID
}

type RemoveUserFromWorkspaceParams struct {
	UserUUID      uuid.UUID
	WorkspaceUUID uuid.UUID
}
