package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"saviour/internal/logger"
	"saviour/internal/service/user"
	"saviour/internal/transport/rest/api"
)

type UserHandler struct {
	log         *slog.Logger
	userService user.Service
}

func NewUserHandler(
	log *slog.Logger,
	userService user.Service,
) *UserHandler {
	return &UserHandler{
		log:         log,
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(
	_ http.ResponseWriter,
	r *http.Request,
	request api.CreateUserRequestObject,
) (api.CreateUserResponseObject, error) {
	usr, err := h.userService.CreateUser(
		r.Context(),
		user.CreateUserParams{
			Username: request.Body.Username,
			Password: request.Body.Password,
			IsAdmin:  request.Body.IsAdmin,
		},
	)
	if err != nil {
		if errors.Is(err, user.ErrUserAlreadyExists) {
			return nil, api.NewErrorResponse(
				http.StatusBadRequest,
				api.UserAlreadyExistsErrorType,
				"User already exists, try a different username",
			)
		}

		h.log.ErrorContext(
			r.Context(),
			"unexpected error calling user service on CreateUser request",
			logger.ErrorAttr(err),
		)

		return nil, api.DefaultInternalError
	}

	return api.UserResponse{
		UUID:      usr.UUID,
		CreatedAt: usr.CreatedAt,
		Username:  usr.Username,
		IsAdmin:   usr.IsAdmin,
	}, nil
}

func (h *UserHandler) DeactivateUser(
	w http.ResponseWriter,
	r *http.Request,
	request api.DeactivateUserRequestObject,
) (api.DeactivateUserResponseObject, error) {
	err := h.userService.DeactivateUser(r.Context(), request.UserUUID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, api.NewErrorResponse(
				http.StatusNotFound,
				api.EntityNotFoundErrorType,
				"User not found",
			)
		}

		return nil, api.DefaultInternalError
	}

	return api.EmptyResponse{}, nil
}
