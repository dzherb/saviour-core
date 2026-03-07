package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"saviour/internal/logger"
	"saviour/internal/model"
	"saviour/internal/service/auth"
	"saviour/internal/service/workspace"
	"saviour/internal/transport/rest/api"
)

type WorkspaceHandler struct {
	log              *slog.Logger
	workspaceService workspace.Service
}

func NewWorkspaceHandler(
	log *slog.Logger,
	workspaceService workspace.Service,
) *WorkspaceHandler {
	return &WorkspaceHandler{
		log:              log,
		workspaceService: workspaceService,
	}
}

func (h *WorkspaceHandler) CreateWorkspace(
	_ http.ResponseWriter,
	r *http.Request,
	request api.CreateWorkspaceRequestObject,
) (api.CreateWorkspaceResponseObject, error) {
	token := auth.MustGetTokenFromContext(r.Context())

	worksp, err := h.workspaceService.CreateWorkspace(
		r.Context(),
		workspace.CreateWorkspaceParams{
			Name:       request.Body.Name,
			AuthorUUID: token.UserUUID,
		},
	)
	if err == nil {
		return mapWorkspace(worksp), nil
	}

	h.log.ErrorContext(
		r.Context(),
		"unexpected error calling workspace service "+
			"on CreateWorkspace request",
		logger.ErrorAttr(err),
	)

	return nil, api.DefaultInternalError
}

func (h *WorkspaceHandler) UpdateWorkspace(
	_ http.ResponseWriter,
	r *http.Request,
	request api.UpdateWorkspaceRequestObject,
) (api.UpdateWorkspaceResponseObject, error) {
	worksp, err := h.workspaceService.UpdateWorkspace(
		r.Context(),
		request.WorkspaceUUID,
		workspace.UpdateWorkspaceParams{
			Name: request.Body.Name,
		},
	)
	if err == nil {
		return mapWorkspace(worksp), nil
	}

	if errors.Is(err, workspace.ErrWorkspaceNotFound) {
		return nil, api.NewErrorResponse(
			http.StatusNotFound,
			api.EntityNotFoundErrorType,
			"Workspace not found",
		)
	}

	h.log.ErrorContext(
		r.Context(),
		"unexpected error calling workspace service "+
			"on UpdateWorkspace request",
		logger.ErrorAttr(err),
	)

	return nil, api.DefaultInternalError
}

func (h *WorkspaceHandler) AddUserToWorkspace(
	_ http.ResponseWriter,
	r *http.Request,
	request api.AddUserToWorkspaceRequestObject,
) (api.AddUserToWorkspaceResponseObject, error) {
	err := h.workspaceService.AddUserToWorkspace(
		r.Context(),
		request.Body.UserUUID,
		request.WorkspaceUUID,
	)
	if err == nil {
		return api.EmptyResponse{}, nil
	}

	if errors.Is(err, workspace.ErrWorkspaceOrUserNotFound) {
		return nil, api.NewErrorResponse(
			http.StatusNotFound,
			api.EntityNotFoundErrorType,
			"User or workspace does not exist",
		)
	}

	h.log.ErrorContext(
		r.Context(),
		"unexpected error calling workspace service "+
			"on AddUserToWorkspace request",
		logger.ErrorAttr(err),
	)

	return nil, api.DefaultInternalError
}

func (h *WorkspaceHandler) RemoveUserFromWorkspace(
	_ http.ResponseWriter,
	r *http.Request,
	request api.RemoveUserFromWorkspaceRequestObject,
) (api.RemoveUserFromWorkspaceResponseObject, error) {
	err := h.workspaceService.RemoveUserFromWorkspace(
		r.Context(),
		request.UserUUID,
		request.WorkspaceUUID,
	)

	if err == nil || errors.Is(err, workspace.ErrWorkspaceUserNotFound) {
		return api.EmptyResponse{}, nil
	}

	h.log.ErrorContext(
		r.Context(),
		"unexpected error calling workspace service "+
			"on RemoveUserFromWorkspace request",
		logger.ErrorAttr(err),
	)

	return nil, api.DefaultInternalError
}

func mapWorkspace(worksp *model.Workspace) *api.WorkspaceResponse {
	return &api.WorkspaceResponse{
		UUID:      worksp.UUID,
		CreatedAt: worksp.CreatedAt,
		Name:      worksp.Name,
	}
}
