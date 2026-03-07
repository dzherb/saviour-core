package handler

import (
	"saviour/internal/transport/rest/api"
)

type APIHandler struct {
	*PingHandler
	*AuthHandler
	*UserHandler
	*WorkspaceHandler
	*SecretHandler
}

var _ api.HandlerInterface = new(APIHandler)

func NewAPIHandler(
	ping *PingHandler,
	auth *AuthHandler,
	user *UserHandler,
	workspace *WorkspaceHandler,
	secret *SecretHandler,
) *APIHandler {
	return &APIHandler{
		PingHandler:      ping,
		AuthHandler:      auth,
		UserHandler:      user,
		WorkspaceHandler: workspace,
		SecretHandler:    secret,
	}
}
