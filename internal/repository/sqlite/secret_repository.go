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

type SecretRepository struct {
	baseRepository
}

func NewSecretRepository(db *sql.DB) *SecretRepository {
	return &SecretRepository{newRepository(db)}
}

func (s SecretRepository) CreateSecret(
	ctx context.Context,
	params repository.CreateSecretParams,
) (*model.Secret, error) {
	if params.UUID == uuid.Nil {
		params.UUID = uuid.Must(uuid.NewV7())
	}

	secret, err := s.txAwareQueries(ctx).CreateSecret(
		ctx,
		sqlitequery.CreateSecretParams{
			UUID:           sqlc.UUID(params.UUID),
			Name:           params.Name,
			ValueEncrypted: params.ValueEncrypted,
			AuthorUUID:     sqlc.UUID(params.AuthorUUID),
			WorkspaceUUID:  sqlc.UUID(params.WorkspaceUUID),
		},
	)
	if err != nil {
		if isForeignKeyViolation(err) {
			return nil, errors.Join(repository.ErrForeignKeyViolation, err)
		}

		return nil, err
	}

	return mapSecret(&secret), nil
}

func (s SecretRepository) UpdateSecret(
	ctx context.Context,
	secretUUID uuid.UUID,
	params repository.UpdateSecretParams,
) (*model.Secret, error) {
	secret, err := s.txAwareQueries(ctx).UpdateSecret(
		ctx,
		sqlitequery.UpdateSecretParams{
			UUID:           sqlc.UUID(secretUUID),
			Name:           params.Name,
			ValueEncrypted: params.ValueEncrypted,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(repository.ErrSecretNotFound, err)
		}

		return nil, err
	}

	return mapSecret(&secret), nil
}

func (s SecretRepository) RenameSecret(
	ctx context.Context,
	secretUUID uuid.UUID,
	params repository.RenameSecretParams,
) (*model.Secret, error) {
	secret, err := s.txAwareQueries(ctx).RenameSecret(
		ctx,
		sqlitequery.RenameSecretParams{
			UUID: sqlc.UUID(secretUUID),
			Name: params.Name,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(repository.ErrSecretNotFound, err)
		}

		return nil, err
	}

	return mapSecret(&secret), nil
}

func mapSecret(secret *sqlitequery.Secret) *model.Secret {
	return &model.Secret{
		UUID:       uuid.UUID(secret.UUID),
		CreatedAt:  secret.CreatedAt.Time(),
		Name:       secret.Name,
		AuthorUUID: uuid.UUID(secret.AuthorUUID),
	}
}
