package secret

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"saviour/internal/model"
	"saviour/internal/repository"
	"saviour/pkg/secret"
)

type ServiceImpl struct {
	secretKey  secret.Secret[[]byte]
	secretRepo repository.SecretRepository
}

func NewService(
	secretKey secret.Secret[[]byte],
	secretRepo repository.SecretRepository,
) *ServiceImpl {
	return &ServiceImpl{
		secretKey:  secretKey,
		secretRepo: secretRepo,
	}
}

func (s ServiceImpl) CreateSecret(
	ctx context.Context,
	params CreateSecretParams,
) (*model.Secret, error) {
	valueEncrypted, err := encrypt(
		s.secretKey.UnsafeValue(),
		[]byte(params.Value),
	)
	if err != nil {
		return nil, err
	}

	secr, err := s.secretRepo.CreateSecret(
		ctx,
		repository.CreateSecretParams{
			Name:           params.Name,
			AuthorUUID:     params.AuthorUUID,
			WorkspaceUUID:  params.WorkspaceUUID,
			ValueEncrypted: valueEncrypted,
		},
	)
	if err != nil {
		if errors.Is(err, repository.ErrForeignKeyViolation) {
			// The author most likely exists at this point,
			// so this interpretation is reasonable enough
			return nil, ErrWorkspaceNotFound
		}

		return nil, err
	}

	return secr, nil
}

func (s ServiceImpl) UpdateSecret(
	ctx context.Context,
	secretUUID uuid.UUID,
	params UpdateSecretParams,
) (*model.Secret, error) {
	valueEncrypted, err := encrypt(
		s.secretKey.UnsafeValue(),
		[]byte(params.Value),
	)
	if err != nil {
		return nil, err
	}

	secr, err := s.secretRepo.UpdateSecret(
		ctx,
		secretUUID,
		repository.UpdateSecretParams{
			Name:           params.Name,
			ValueEncrypted: valueEncrypted,
		},
	)
	if err != nil {
		if errors.Is(err, repository.ErrSecretNotFound) {
			return nil, ErrSecretNotFound
		}

		return nil, err
	}

	return secr, nil
}
