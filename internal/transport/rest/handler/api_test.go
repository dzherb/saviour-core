package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"saviour/internal/logger"
	"saviour/internal/service/auth"
	"saviour/internal/transport/rest/api"
)

type APITestSuite struct {
	suite.Suite

	spec   *openapi3.T
	router routers.Router

	AuthToken *auth.TokenParsed
}

func (suite *APITestSuite) SetupSuite() {
	spec, err := api.GetSwagger()
	if err != nil {
		panic("parsing swagger spec: " + err.Error())
	}

	// Disable host validation
	spec.Servers = nil

	suite.spec = spec

	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		panic("creating router: " + err.Error())
	}

	suite.router = router
}

func (suite *APITestSuite) PreparedRecorderAndRequest(
	method, path string,
	body map[string]any,
	userRoles []auth.Role,
) (*httptest.ResponseRecorder, *http.Request) {
	t := suite.T()
	t.Helper()

	bodyBytes, err := json.Marshal(body)
	suite.Require().NoError(err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))

	r.Header.Set("Content-Type", "application/json")

	suite.AuthToken = &auth.TokenParsed{
		UserUUID:    uuid.New(),
		SessionUUID: uuid.New(),
		Roles:       userRoles,
	}

	r = r.WithContext(
		auth.ContextWithToken(r.Context(), suite.AuthToken),
	)

	route, pathParams, err := suite.router.FindRoute(r)
	suite.Require().NoError(err)

	suite.T().Cleanup(func() {
		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: &openapi3filter.RequestValidationInput{
				Request:    r,
				Route:      route,
				PathParams: pathParams,
				Options: &openapi3filter.Options{
					AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
				},
			},
			Status: w.Result().StatusCode,
			Header: w.Result().Header,
			Options: &openapi3filter.Options{
				IncludeResponseStatus: true,
			},
		}
		responseValidationInput.SetBodyBytes(w.Body.Bytes())

		err = openapi3filter.ValidateResponse(
			context.Background(),
			responseValidationInput,
		)
		suite.Require().NoError(err)
	})

	return w, r
}

func (suite *APITestSuite) ServeAPI(
	w http.ResponseWriter,
	r *http.Request,
	hi api.HandlerInterface,
) {
	h := api.NewHTTPHandler(
		logger.Noop,
		hi,
	)

	h.ServeHTTP(w, r)
}

func TestAPISuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(APITestSuite))
}
