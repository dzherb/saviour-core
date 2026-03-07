package secret

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"saviour/internal/model"
)

type Service interface {
	CreateSecret(
		ctx context.Context,
		params CreateSecretParams,
	) (*model.Secret, error)
	UpdateSecret(
		ctx context.Context,
		secretUUID uuid.UUID,
		params UpdateSecretParams,
	) (*model.Secret, error)
}

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrSecretNotFound    = errors.New("secret not found")
)

type CreateSecretParams struct {
	AuthorUUID    uuid.UUID
	WorkspaceUUID uuid.UUID
	Name          string
	Value         string
}

type UpdateSecretParams struct {
	Name  string
	Value string
}
