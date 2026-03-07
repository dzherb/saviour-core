package sqliterepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"saviour/internal/model"
	"saviour/internal/repository"
	"saviour/internal/repository/sqlite/sqlc"
	sqlitequery "saviour/internal/repository/sqlite/sqlc/gen"
)

type WorkspaceRepository struct {
	baseRepository
}

func NewWorkspaceRepository(db *sql.DB) *WorkspaceRepository {
	return &WorkspaceRepository{newRepository(db)}
}

func (r *WorkspaceRepository) CreateWorkspace(
	ctx context.Context,
	params repository.CreateWorkspaceParams,
) (*model.Workspace, error) {
	if params.UUID == uuid.Nil {
		params.UUID = uuid.Must(uuid.NewV7())
	}

	workspace, err := r.txAwareQueries(ctx).
		CreateWorkspace(
			ctx,
			sqlitequery.CreateWorkspaceParams{
				UUID:       sqlc.UUID(params.UUID),
				Name:       params.Name,
				AuthorUUID: sqlc.UUID(params.AuthorUUID),
			},
		)
	if err != nil {
		return nil, err
	}

	return mapWorkspace(&workspace), nil
}

func (r *WorkspaceRepository) UpdateWorkspace(
	ctx context.Context,
	workspaceUUID uuid.UUID,
	params repository.UpdateWorkspaceParams,
) (*model.Workspace, error) {
	workspace, err := r.txAwareQueries(ctx).
		UpdateWorkspace(
			ctx,
			sqlitequery.UpdateWorkspaceParams{
				UUID: sqlc.UUID(workspaceUUID),
				Name: params.Name,
			},
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(repository.ErrWorkspaceNotFound, err)
		}

		return nil, err
	}

	return mapWorkspace(&workspace), nil
}

func (r *WorkspaceRepository) AddUserToWorkspace(
	ctx context.Context,
	userUUID uuid.UUID,
	workspaceUUID uuid.UUID,
) error {
	err := r.txAwareQueries(ctx).
		AddUserToWorkspace(
			ctx,
			sqlitequery.AddUserToWorkspaceParams{
				UserUUID:      sqlc.UUID(userUUID),
				WorkspaceUUID: sqlc.UUID(workspaceUUID),
			},
		)
	if err != nil {
		if isForeignKeyViolation(err) {
			return errors.Join(repository.ErrForeignKeyViolation, err)
		}

		return err
	}

	return nil
}

func (r *WorkspaceRepository) RemoveUserFromWorkspace(
	ctx context.Context,
	userUUID uuid.UUID,
	workspaceUUID uuid.UUID,
) error {
	rowsAffected, err := r.txAwareQueries(ctx).
		RemoveUserFromWorkspace(
			ctx,
			sqlitequery.RemoveUserFromWorkspaceParams{
				UserUUID:      sqlc.UUID(userUUID),
				WorkspaceUUID: sqlc.UUID(workspaceUUID),
			},
		)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrWorkspaceUserNotFound
	}

	return nil
}

func (r *WorkspaceRepository) DoesUserBelongToWorkspace(
	ctx context.Context,
	userUUID uuid.UUID,
	workspaceUUID uuid.UUID,
) (bool, error) {
	return r.txAwareQueries(ctx).DoesUserBelongToWorkspace(
		ctx,
		sqlitequery.DoesUserBelongToWorkspaceParams{
			UserUUID:      sqlc.UUID(userUUID),
			WorkspaceUUID: sqlc.UUID(workspaceUUID),
		},
	)
}

func mapWorkspace(workspace *sqlitequery.Workspace) *model.Workspace {
	return &model.Workspace{
		UUID:       uuid.UUID(workspace.UUID),
		CreatedAt:  workspace.CreatedAt.Time(),
		Name:       workspace.Name,
		AuthorUUID: uuid.UUID(workspace.AuthorUUID),
	}
}
