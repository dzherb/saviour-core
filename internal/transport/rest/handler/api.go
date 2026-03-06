package handler

import (
	"saviour/internal/transport/rest/api"
)

type APIHandler struct {
	*PingHandler
	*AuthHandler
}

var _ api.HandlerInterface = new(APIHandler)

func NewAPIHandler(
	ping *PingHandler,
	auth *AuthHandler,
) *APIHandler {
	return &APIHandler{
		PingHandler: ping,
		AuthHandler: auth,
	}
}
