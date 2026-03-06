package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"saviour/internal/logger"
)

func WriteResponse(
	log *slog.Logger,
	w http.ResponseWriter,
	r *http.Request,
	obj any,
	status int32,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(status))

	if obj == nil {
		return
	}

	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		log.ErrorContext(
			r.Context(),
			"failed to write response",
			logger.ErrorAttr(err),
		)
	}
}

type empty struct{}
