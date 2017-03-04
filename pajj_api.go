package main

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

type authorizeRequest struct {
	SessionToken string `json:"session_token"`
}

func main() {
	print("starting server")
	mux := http.NewServeMux()

	mux.HandleFunc("/authorize", authorize)

	err := http.ListenAndServe(":8118", mux)
	log.Fatal(err)
}

func authorize(w http.ResponseWriter, r *http.Request) {
	authReq := authorizeRequest{}
	jsonText, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(jsonText, &authReq)
	w.WriteHeader(200)

	fmt.Fprintln(w, "Hello browser")
}