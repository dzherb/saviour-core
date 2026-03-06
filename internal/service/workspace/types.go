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

type CreateWorkspaceParams struct {
	Name       string
	AuthorUUID uuid.UUID
}

type UpdateWorkspaceParams repository.UpdateWorkspaceParams

type AddUserToWorkspaceParams repository.AddUserToWorkspaceParams

type RemoveUserFromWorkspaceParams repository.RemoveUserFromWorkspaceParams
