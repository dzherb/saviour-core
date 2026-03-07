package workspace

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"saviour/internal/model"
	"saviour/internal/repository"
)

var (
	ErrWorkspaceNotFound       = errors.New("workspace not found")
	ErrWorkspaceOrUserNotFound = errors.New("workspace or user not found")
	ErrWorkspaceUserNotFound   = errors.New("workspace user not found")
)

type Service interface {
	CreateWorkspace(
		ctx context.Context,
		params CreateWorkspaceParams,
	) (*model.Workspace, error)
	UpdateWorkspace(
		ctx context.Context,
		workspaceUUID uuid.UUID,
		params UpdateWorkspaceParams,
	) (*model.Workspace, error)
	AddUserToWorkspace(
		ctx context.Context,
		userUUID uuid.UUID,
		workspaceUUID uuid.UUID,
	) error
	RemoveUserFromWorkspace(
		ctx context.Context,
		userUUID uuid.UUID,
		workspaceUUID uuid.UUID,
	) error
}

type CreateWorkspaceParams struct {
	Name       string
	AuthorUUID uuid.UUID
}

type UpdateWorkspaceParams repository.UpdateWorkspaceParams
