package acl

import (
	"context"
	"slices"

	"saviour/internal/repository"
	"saviour/internal/service/auth"
)

type ServiceImpl struct {
	workspaceRepo repository.WorkspaceRepository
}

func NewService(
	workspaceRepo repository.WorkspaceRepository,
) *ServiceImpl {
	return &ServiceImpl{
		workspaceRepo: workspaceRepo,
	}
}

func (s ServiceImpl) CanAccessWorkspaceResource(
	ctx context.Context,
	token *auth.TokenParsed,
	params CanAccessWorkspaceResourceParams,
) (bool, error) {
	if hasAdminRole(token) {
		return true, nil
	}

	return s.workspaceRepo.DoesUserBelongToWorkspace(
		ctx,
		token.UserUUID,
		params.WorkspaceUUID,
	)
}

func hasAdminRole(token *auth.TokenParsed) bool {
	return slices.Contains(token.Roles, auth.AdminRole)
}
