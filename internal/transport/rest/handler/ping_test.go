package handler_test

import (
	"encoding/json"
	"net/http"

	"saviour/internal/transport/rest/handler"
)

func (suite *APITestSuite) TestPing_OK() {
	w, r := suite.PreparedRecorderAndRequest(http.MethodGet, "/ping", nil)

	h := handler.APIHandler{
		PingHandler: handler.NewPingHandler("test"),
	}

	suite.ServeAPI(w, r, h)

	suite.Require().Equal(http.StatusOK, w.Code)

	response := make(map[string]any)
	err := json.Unmarshal(w.Body.Bytes(), &response)

	suite.Require().NoError(err)

	suite.Require().Contains(response, "instance")
	suite.Equal(response["instance"], "test")
}
