package main

import (
	"net/http/httptest"
	"net/http"
	"encoding/json"
	"strings"
	"testing"
	"fmt"
	"reflect"
)

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

type RPCMock struct {
	amqpReq, amqpRes string
	t                *testing.T
}

func (mock RPCMock) SendRequest(method string, body []byte) chan []byte {
	jsonCmpOk, err := AreEqualJSON(method, mock.amqpReq)
	if err != nil {
		mock.t.Errorf("JSON check: %s", err)
	}

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
	jsonCmpOk, err := AreEqualJSON(response.Body, `{"session_token": "DEF"}`)
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