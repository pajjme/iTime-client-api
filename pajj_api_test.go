package main

import (
	"github.com/pajjme/iTime-client-api/apiutil"
	"net/http/httptest"
	"strings"
	"testing"
	"io/ioutil"
)

type RPCMock struct {
	amqpReq, amqpRes string
	t                *testing.T
}

func (mock RPCMock) SendRequest(method string, body []byte) chan []byte {
	apiutil.AssertEqualJSON(mock.amqpReq, string(body))

	response := make(chan []byte, 1)
	response <- []byte(mock.amqpRes)
	return response
}

func TestAuthorize(t *testing.T) {
	request := httptest.NewRequest("POST", "/v1/authorize", strings.NewReader(`{"auth_code": "ABC", "junk": 123}`))
	response := httptest.NewRecorder()
	rpc := RPCMock{`{"auth_code": "ABC"}`, `{"session_token": "DEF", "code": 0}`, t}

	apiutil.Authorize(response, request, rpc)
	restRes := `{"session_token": "DEF"}`
	body, err := ioutil.ReadAll(response.Body)
	apiutil.CheckError(err)
	apiutil.AssertEqualJSON(string(body), restRes)
}

