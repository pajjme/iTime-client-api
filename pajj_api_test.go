package main

import (
	"net/http/httptest"
	"net/http"
	"encoding/json"
	"strings"
	"testing"
)

func testAuthorize(t *testing.T) {

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