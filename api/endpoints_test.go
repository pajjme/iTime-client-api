package api

import (
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

type RPCMock struct {
	t                *testing.T
	amqpReq, amqpRes string
}

func (mock RPCMock) SendRequest(method string, body []byte) chan []byte {
	AssertEqualJSON(mock.amqpReq, string(body))

	response := make(chan []byte, 1)
	response <- []byte(mock.amqpRes)
	return response
}

func TestAuthorize(t *testing.T) {
	request := httptest.NewRequest("POST", "/v1/authorize",
		strings.NewReader(`{"auth_code": "ABC", "junk": 123}`))
	response := httptest.NewRecorder()
	rpc := RPCMock{t, `{"auth_code": "ABC"}`, `{"session_token": "DEF", "code": 0}`}
	Authorize(response, request, rpc)

	assert.Equal(t, 200, response.Code)
	assert.Regexp(t, "sessionToken=[^;]+", response.HeaderMap.Get("Set-Cookie"))
}

func TestAuthorizeInvalidJson(t *testing.T) {
	request := httptest.NewRequest("POST", "/v1/authorize",
		strings.NewReader(`{`)) // No '}' at end
	response := httptest.NewRecorder()
	rpc := RPCMock{t, ``, ``}
	Authorize(response, request, rpc)

	assert.Equal(t, 400, response.Code)
}

func TestAuthorizeUnauthorized(t *testing.T) {
	request := httptest.NewRequest("POST", "/v1/authorize",
		strings.NewReader(`{"auth_code": "ABC"}`))
	response := httptest.NewRecorder()
	rpc := RPCMock{t, `{"auth_code": "ABC"}`, `{"error": "...", "code": 1}`}
	Authorize(response, request, rpc)

	assert.Equal(t, 401, response.Code)
}

func TestAuthorizeInternalServerError(t *testing.T) {
	request := httptest.NewRequest("POST", "/v1/authorize",
		strings.NewReader(`{"auth_code": "ABC"}`))
	response := httptest.NewRecorder()
	rpc := RPCMock{t, `{"auth_code": "ABC"}`, `{"error": "...", "code": 2}`}
	Authorize(response, request, rpc)

	assert.Equal(t, 500, response.Code)
}