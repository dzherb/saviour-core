package acl

import (
	"context"

	"github.com/google/uuid"

	"saviour/internal/service/auth"
)

type Service interface {
	CanAccessWorkspaceResource(
		ctx context.Context,
		token *auth.TokenParsed,
		params CanAccessWorkspaceResourceParams,
	) (bool, error)
}

type CanAccessWorkspaceResourceParams struct {
	WorkspaceUUID uuid.UUID
	ResourceUUID  uuid.UUID
}
