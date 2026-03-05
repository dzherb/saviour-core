package sqliterepo

import (
	"context"
	"database/sql"

	"saviour/internal/model"
	"saviour/internal/repository"
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
	user, err := r.txAwareQueries(ctx).
		CreateUser(
			ctx,
			sqlitequery.CreateUserParams(params),
		)

	if err != nil {
		return nil, err
	}

	return mapUser(&user), nil
}

func mapUser(user *sqlitequery.User) *model.User {
	return &model.User{
		UUID:      user.UUID,
		CreatedAt: user.CreatedAt.Time(),
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
	}
}
