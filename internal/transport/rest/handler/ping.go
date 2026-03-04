package handler

import (
	"net/http"
	"time"

	"saviour/internal/transport/rest/api"
)

type PingHandler struct {
	appInstance string
}

func NewPingHandler(instance string) *PingHandler {
	return &PingHandler{
		appInstance: instance,
	}
}

func (h *PingHandler) Ping(
	http.ResponseWriter,
	*http.Request,
	api.PingRequestObject,
) (api.PingResponseObject, error) {
	return api.PingResponse{
		Instance:   h.appInstance,
		ServerTime: time.Now().UTC(),
	}, nil
}
