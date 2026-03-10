package acl_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"saviour/internal/service/acl"
	"saviour/internal/service/auth"
	"saviour/internal/testkit/testmock"
)

func TestService_CanAccessWorkspaceResource_UserBelongsToWorkspace(
	t *testing.T,
) {
	t.Parallel()

	workspaceRepo := testmock.NewMockWorkspaceRepository(t)

	userUUID := uuid.New()
	workspaceUUID := uuid.New()

	workspaceRepo.EXPECT().DoesUserBelongToWorkspace(
		mock.Anything,
		userUUID,
		workspaceUUID,
	).Return(true, nil)

	service := acl.NewService(workspaceRepo)

	hasAccess, err := service.CanAccessWorkspaceResource(
		context.Background(),
		&auth.TokenParsed{
			UserUUID:    userUUID,
			SessionUUID: uuid.New(),
			Roles:       []auth.Role{},
		},
		acl.CanAccessWorkspaceResourceParams{
			WorkspaceUUID: workspaceUUID,
			ResourceUUID:  uuid.New(),
		},
	)

	require.NoError(t, err)
	require.True(t, hasAccess)
}

func TestService_CanAccessWorkspaceResource_UserIsAdmin(t *testing.T) {
	t.Parallel()

	workspaceRepo := testmock.NewMockWorkspaceRepository(t)
	service := acl.NewService(workspaceRepo)

	hasAccess, err := service.CanAccessWorkspaceResource(
		context.Background(),
		&auth.TokenParsed{
			UserUUID:    uuid.New(),
			SessionUUID: uuid.New(),
			Roles:       []auth.Role{auth.AdminRole},
		},
		acl.CanAccessWorkspaceResourceParams{
			WorkspaceUUID: uuid.New(),
			ResourceUUID:  uuid.New(),
		},
	)

	require.NoError(t, err)
	require.True(t, hasAccess)
}
