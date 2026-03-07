package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"saviour/internal/logger"
	"saviour/internal/model"
	"saviour/internal/service/acl"
	"saviour/internal/service/auth"
	"saviour/internal/service/secret"
	"saviour/internal/transport/rest/api"
)

type SecretHandler struct {
	log           *slog.Logger
	aclService    acl.Service
	secretService secret.Service
}

func NewSecretHandler(
	log *slog.Logger,
	aclService acl.Service,
	secretService secret.Service,
) *SecretHandler {
	return &SecretHandler{
		log:           log,
		aclService:    aclService,
		secretService: secretService,
	}
}

func (h *SecretHandler) CreateSecret(
	_ http.ResponseWriter,
	r *http.Request,
	request api.CreateSecretRequestObject,
) (api.CreateSecretResponseObject, error) {
	token := auth.MustGetTokenFromContext(r.Context())

	ok, err := h.aclService.CanAccessWorkspaceResource(
		r.Context(),
		token,
		acl.CanAccessWorkspaceResourceParams{
			WorkspaceUUID: request.WorkspaceUUID,
		},
	)
	if err != nil {
		h.log.ErrorContext(
			r.Context(),
			"unexpected error calling acl service on CreateSecret request",
			logger.ErrorAttr(err),
		)

		return nil, api.DefaultInternalError
	}

	if !ok {
		return nil, api.WorkspaceResourceForbiddenError
	}

	secr, err := h.secretService.CreateSecret(
		r.Context(),
		secret.CreateSecretParams{
			AuthorUUID:    token.UserUUID,
			WorkspaceUUID: request.WorkspaceUUID,
			Name:          request.Body.Name,
			Value:         request.Body.Value,
		},
	)
	if err != nil {
		if errors.Is(err, secret.ErrWorkspaceNotFound) {
			return nil, api.NewErrorResponse(
				http.StatusNotFound,
				api.EntityNotFoundErrorType,
				"Workspace not found",
			)
		}

		h.log.ErrorContext(
			r.Context(),
			"unexpected error calling secret service on CreateSecret request",
			logger.ErrorAttr(err),
		)

		return nil, api.DefaultInternalError
	}

	return mapSecret(secr), nil
}

func (h *SecretHandler) UpdateSecret(
	_ http.ResponseWriter,
	r *http.Request,
	request api.UpdateSecretRequestObject,
) (api.UpdateSecretResponseObject, error) {
	token := auth.MustGetTokenFromContext(r.Context())

	ok, err := h.aclService.CanAccessWorkspaceResource(
		r.Context(),
		token,
		acl.CanAccessWorkspaceResourceParams{
			WorkspaceUUID: request.WorkspaceUUID,
		},
	)
	if err != nil {
		h.log.ErrorContext(
			r.Context(),
			"unexpected error calling acl service on UpdateSecret request",
			logger.ErrorAttr(err),
		)

		return nil, api.DefaultInternalError
	}

	if !ok {
		return nil, api.WorkspaceResourceForbiddenError
	}

	secr, err := h.secretService.UpdateSecret(
		r.Context(),
		request.SecretUUID,
		secret.UpdateSecretParams{
			Name:  request.Body.Name,
			Value: request.Body.Value,
		},
	)
	if err != nil {
		if errors.Is(err, secret.ErrSecretNotFound) {
			return nil, api.NewErrorResponse(
				http.StatusNotFound,
				api.EntityNotFoundErrorType,
				"Secret not found",
			)
		}

		h.log.ErrorContext(
			r.Context(),
			"unexpected error calling secret service on UpdateSecret request",
			logger.ErrorAttr(err),
		)

		return nil, api.DefaultInternalError
	}

	return mapSecret(secr), nil
}

func mapSecret(s *model.Secret) *api.SecretResponse {
	return &api.SecretResponse{
		UUID:       s.UUID,
		CreatedAt:  s.CreatedAt,
		Name:       s.Name,
		AuthorUUID: s.AuthorUUID,
	}
}
