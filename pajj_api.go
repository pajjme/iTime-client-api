package main

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

const ApiVersion string = "/v1"

type authorizeRequest struct {
	SessionToken string `json:"session_token"`
}

func main() {
	print("starting server")

	err := http.ListenAndServe(":8118", mainHandler())
	log.Fatal(err)
}

func mainHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(ApiVersion + "/authorize", authorize)
	mux.HandleFunc(ApiVersion + "/stats", stats)
	return mux
}

func authorize(w http.ResponseWriter, r *http.Request) {
	authReq := authorizeRequest{}
	jsonText, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(jsonText, &authReq)
	w.WriteHeader(200)

	fmt.Fprintln(w, "Hello browser")
}

func stats(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	from, ok1 := params["from"]
	to, ok2 := params["to"]

	if !ok1 || !ok2 {
		w.WriteHeader(400) // Bad Request
		return
	}
	_, _ = from, to

	// TODO:  send RPC call, and respond on HTTP request
	fmt.Fprintln(w, "{}")
}