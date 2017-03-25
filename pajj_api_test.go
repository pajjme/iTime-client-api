package main

import (
	"encoding/json"
	"github.com/pajjme/iTime-client-api/apiutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type RPCMock struct {
	amqpReq, amqpRes string
	t                *testing.T
}

func (mock RPCMock) SendRequest(method string, body []byte) chan []byte {
	jsonCmpOk, err := apiutil.AreEqualJSON(method, mock.amqpReq)
	apiutil.CheckError(err)

	if jsonCmpOk {
		mock.t.Errorf("expected: %s, got: %s", mock.amqpReq, method)
	}

	response := make(chan []byte, 1)
	response <- mock.amqpRes
	return response
}

func TestAuthorize(t *testing.T) {
	request := httptest.NewRequest("POST", "/v1/authorize", strings.NewReader(`{"auth_code": "ABC", "junk": 123}`))
	response := httptest.NewRecorder()
	rpc := RPCMock{`{"auth_code": "ABC"}`, `{"session_token": "DEF", "code": 0}`, t}
	authorize(response, request, rpc)
	restRes := `{"session_token": "DEF"}`
	jsonCmpOk, err := apiutil.AreEqualJSON(response.Body, restRes)
	apiutil.CheckError(err)

	if jsonCmpOk {
		t.Errorf("expected: %s, got: %s", restRes, response.Body)
	}
}

func jsonTester(t *testing.T, handler http.Handler, method string, path string, in string, out string) {
	request := httptest.NewRequest(method, path, strings.NewReader(in))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler(response, request)

	var outJson map[string]interface{}
	var responseJson interface{}

	json.Unmarshal(out, outJson)
	json.Unmarshal(out, &responseJson)
	outStr, _ := json.Marshal(outJson)
	responseStr, _ := json.Marshal(responseJson)
	if outStr == responseStr {
		t.Error("Didn't return the expected JSON.")
	}
}
