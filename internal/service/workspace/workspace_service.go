package workspace

import (
	"context"
	"errors"
	"log/slog"

	"saviour/internal/model"
	"saviour/internal/repository"
)

type ServiceImpl struct {
	log           *slog.Logger
	workspaceRepo repository.WorkspaceRepository
}

func NewService(
	log *slog.Logger,
	workspaceRepo repository.WorkspaceRepository,
) *ServiceImpl {
	return &ServiceImpl{
		log:           log,
		workspaceRepo: workspaceRepo,
	}
}

func (s ServiceImpl) CreateWorkspace(
	ctx context.Context,
	params CreateWorkspaceParams,
) (*model.Workspace, error) {
	return s.workspaceRepo.CreateWorkspace(
		ctx,
		repository.CreateWorkspaceParams{
			Name:       params.Name,
			AuthorUUID: params.AuthorUUID,
		},
	)
}

func (s ServiceImpl) UpdateWorkspace(
	ctx context.Context,
	params UpdateWorkspaceParams,
) (*model.Workspace, error) {
	workspace, err := s.workspaceRepo.UpdateWorkspace(
		ctx,
		repository.UpdateWorkspaceParams(params),
	)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return nil, ErrWorkspaceNotFound
		}

		return nil, err
	}

	return workspace, nil
}

func (s ServiceImpl) AddUserToWorkspace(
	ctx context.Context,
	params AddUserToWorkspaceParams,
) error {
	err := s.workspaceRepo.AddUserToWorkspace(
		ctx,
		repository.AddUserToWorkspaceParams(params),
	)
	if err != nil {
		if errors.Is(err, repository.ErrForeignKeyViolation) {
			return ErrWorkspaceOrUserNotFound
		}

		return err
	}

	return nil
}

func (s ServiceImpl) RemoveUserFromWorkspace(
	ctx context.Context,
	params RemoveUserFromWorkspaceParams,
) error {
	err := s.workspaceRepo.RemoveUserFromWorkspace(
		ctx,
		repository.RemoveUserFromWorkspaceParams(params),
	)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceUserNotFound) {
			return ErrWorkspaceUserNotFound
		}

		return err
	}

	return nil
}
