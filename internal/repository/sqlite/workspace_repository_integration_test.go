package sqliterepo_test

import (
	"context"

	"github.com/google/uuid"

	"saviour/internal/repository"
	sqliterepo "saviour/internal/repository/sqlite"
)

func (suite *RepositoryTestSuite) TestWorkspaceRepository_AddUser_FKViolation() {
	workspaceRepo := sqliterepo.NewWorkspaceRepository(suite.DB)

	err := workspaceRepo.AddUserToWorkspace(
		context.Background(),
		uuid.New(),
		uuid.New(),
	)

	suite.ErrorIs(err, repository.ErrForeignKeyViolation)
}

func (suite *RepositoryTestSuite) TestWorkspaceRepository_RemoveUser_NotFound() {
	workspaceRepo := sqliterepo.NewWorkspaceRepository(suite.DB)

	err := workspaceRepo.RemoveUserFromWorkspace(
		context.Background(),
		uuid.New(),
		uuid.New(),
	)

	suite.ErrorIs(err, repository.ErrWorkspaceUserNotFound)
}
