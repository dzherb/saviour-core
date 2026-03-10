package handler_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"saviour/internal/logger"
	"saviour/internal/model"
	"saviour/internal/service/auth"
	"saviour/internal/service/user"
	"saviour/internal/testkit/testmock"
	"saviour/internal/transport/rest/api"
	"saviour/internal/transport/rest/handler"
)

func (suite *APITestSuite) TestUser_CreateUser_OK() {
	w, r := suite.PreparedRecorderAndRequest(
		http.MethodPost,
		"/users",
		map[string]any{
			"username": "test_user",
			"password": "password",
		},
		[]auth.Role{auth.AdminRole},
	)

	userService := testmock.NewMockUserService(suite.T())

	userService.EXPECT().
		CreateUser(
			mock.Anything,
			user.CreateUserParams{
				Username: "test_user",
				Password: "password",
				IsAdmin:  false,
			},
		).
		Return(
			&model.User{
				UUID:     uuid.New(),
				Username: "test_user",
				IsAdmin:  false,
			},
			nil,
		)

	h := handler.APIHandler{
		UserHandler: handler.NewUserHandler(
			logger.Noop,
			userService,
		),
	}

	suite.ServeAPI(w, r, h)

	suite.Require().Equal(http.StatusOK, w.Code)

	response := make(map[string]any)
	err := json.Unmarshal(w.Body.Bytes(), &response)

	suite.Require().NoError(err)

	suite.Equal(response["username"], "test_user")
}

func (suite *APITestSuite) TestUser_CreateUser_UserAlreadyExists() {
	w, r := suite.PreparedRecorderAndRequest(
		http.MethodPost,
		"/users",
		map[string]any{
			"username": "test_user",
			"password": "password",
		},
		[]auth.Role{auth.AdminRole},
	)

	userService := testmock.NewMockUserService(suite.T())

	userService.EXPECT().
		CreateUser(
			mock.Anything,
			user.CreateUserParams{
				Username: "test_user",
				Password: "password",
				IsAdmin:  false,
			},
		).
		Return(
			nil,
			user.ErrUserAlreadyExists,
		)

	h := handler.APIHandler{
		UserHandler: handler.NewUserHandler(
			logger.Noop,
			userService,
		),
	}

	suite.ServeAPI(w, r, h)

	suite.Require().Equal(http.StatusBadRequest, w.Code)
	suite.Contains(w.Body.String(), api.UserAlreadyExistsErrorType)
}

func (suite *APITestSuite) TestUser_DeactivateUser_OK() {
	userUUID := uuid.New()

	w, r := suite.PreparedRecorderAndRequest(
		http.MethodPost,
		"/users/"+userUUID.String()+"/deactivate",
		nil,
		[]auth.Role{auth.AdminRole},
	)

	userService := testmock.NewMockUserService(suite.T())

	userService.EXPECT().
		DeactivateUser(
			mock.Anything,
			userUUID,
		).
		Return(nil)

	h := handler.APIHandler{
		UserHandler: handler.NewUserHandler(
			logger.Noop,
			userService,
		),
	}

	suite.ServeAPI(w, r, h)

	suite.Require().Equal(http.StatusOK, w.Code)
}

func (suite *APITestSuite) TestUser_DeactivateUser_NotFound() {
	userUUID := uuid.New()

	w, r := suite.PreparedRecorderAndRequest(
		http.MethodPost,
		"/users/"+userUUID.String()+"/deactivate",
		nil,
		[]auth.Role{auth.AdminRole},
	)

	userService := testmock.NewMockUserService(suite.T())

	userService.EXPECT().
		DeactivateUser(
			mock.Anything,
			userUUID,
		).
		Return(user.ErrUserNotFound)

	h := handler.APIHandler{
		UserHandler: handler.NewUserHandler(
			logger.Noop,
			userService,
		),
	}

	suite.ServeAPI(w, r, h)

	suite.Require().Equal(http.StatusNotFound, w.Code)
	suite.Require().Contains(w.Body.String(), api.EntityNotFoundErrorType)
}
