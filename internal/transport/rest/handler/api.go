package handler

import (
	"saviour/internal/transport/rest/api"
)

type APIHandler struct {
	*PingHandler
	*AuthHandler
	*UserHandler
}

var _ api.HandlerInterface = new(APIHandler)

func NewAPIHandler(
	ping *PingHandler,
	auth *AuthHandler,
	user *UserHandler,
) *APIHandler {
	return &APIHandler{
		PingHandler: ping,
		AuthHandler: auth,
		UserHandler: user,
	}
}
