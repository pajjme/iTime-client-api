package apiutil

import (
	"encoding/json"
	"net/http"
	"time"
	"fmt"
	"io/ioutil"
	"log"
)

type authorizeRequest struct {
	AuthCode string `json:"auth_code"`
}

func Authorize(w http.ResponseWriter, r *http.Request, rpc RPCaller) {
	authReq := authorizeRequest{}
	jsonText, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(jsonText, &authReq)
	CheckError(err)

	log.Println("Send request to US", authReq)
	rpcRequest, err := json.Marshal(authReq)
	amqpResponse := <-rpc.SendRequest("authorize", rpcRequest)

	// TODO: Use data from amqpResponse to send to client

	w.WriteHeader(200) // HTTP Found

	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   "",
		Expires: time.Now().AddDate(1, 0, 0), // One year ahead
	})

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintln(w, string(amqpResponse))
}

func Stats(w http.ResponseWriter, r *http.Request, rpc RPCaller) {
	println("stttta")
	params := r.URL.Query()
	from, ok1 := params["from"]
	to, ok2 := params["to"]

	if !ok1 || !ok2 {
		w.WriteHeader(400) // Bad Request
		return
	}
	_, _ = from, to

	// TODO:  send RPC call, and respond on HTTP request
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintln(w, "{}")
}
