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

type UserRepository struct {
	baseRepository
}

var _ repository.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{newRepository(db)}
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	params repository.CreateUserParams,
) (*model.User, error) {
	if params.UUID == uuid.Nil {
		params.UUID = uuid.Must(uuid.NewV7())
	}

	user, err := r.txAwareQueries(ctx).
		CreateUser(
			ctx,
			sqlitequery.CreateUserParams{
				UUID:         sqlc.UUID(params.UUID),
				Username:     params.Username,
				PasswordHash: params.PasswordHash,
				IsAdmin:      params.IsAdmin,
			},
		)
	if err != nil {
		return nil, err
	}

	return mapUser(&user), nil
}

func (r *UserRepository) DeactivateUser(
	ctx context.Context,
	uuid uuid.UUID,
) error {
	rowsAffected, err := r.txAwareQueries(ctx).
		DeactivateUser(ctx, sqlc.UUID(uuid))
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*model.User, error) {
	user, err := r.txAwareQueries(ctx).
		GetUserByUsername(
			ctx,
			username,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(repository.ErrUserNotFound, err)
		}

		return nil, err
	}

	return mapUser(&user), nil
}

func mapUser(user *sqlitequery.User) *model.User {
	return &model.User{
		UUID:         uuid.UUID(user.UUID),
		CreatedAt:    user.CreatedAt.Time(),
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		IsAdmin:      user.IsAdmin,
	}
}
