package handler

import (
	"saviour/internal/transport/rest/api"
)

type APIHandler struct {
	*PingHandler
}

var _ api.HandlerInterface = new(APIHandler)

func NewAPIHandler(
	ping *PingHandler,
) *APIHandler {
	return &APIHandler{
		PingHandler: ping,
	}
}
