package apiutil

import (
	"encoding/json"
	"net/http"
	"time"
	"fmt"
	"io/ioutil"
	"log"
)

type httpAuthorizeRequest struct {
	AuthCode string `json:"auth_code"`
}

type httpAuthorizeResponse struct {
	Error string `json:"error,omitempty"`
}

type amqpAuthorizeRequest struct {
	AuthCode string `json:"auth_code"`
}

type amqpAuthorizeResponse struct {
	SessionToken string `json:"session_token"`
	Error        string `json:"error"`
	Code         int `json:"code"`
}

func Authorize(w http.ResponseWriter, r *http.Request, rpc RPCaller) {
	httpReq := httpAuthorizeRequest{}

	// TODO: Neater error handling?
	jsonText, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	err = json.Unmarshal(jsonText, &httpReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	amqpReq := amqpAuthorizeRequest{httpReq.AuthCode}
	strReq, err := json.Marshal(amqpReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	// TODO: Add a timeout
	amqpResponse := <-rpc.SendRequest("authorize", strReq)

	amqpRes := amqpAuthorizeResponse{}
	err = json.Unmarshal(amqpResponse, &amqpRes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	if amqpRes.Code != 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}


	// TODO: Use data from amqpResponse to send to client

	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   amqpRes.SessionToken,
		Expires: time.Now().AddDate(1, 0, 0), // One year ahead
	})

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, )
}

func Stats(w http.ResponseWriter, r *http.Request, rpc RPCaller) {
	println("stttta")
	params := r.URL.Query()
	from, ok1 := params["from"]
	to, ok2 := params["to"]

	if !ok1 || !ok2 {
		w.WriteHeader(400) // Bad Request
		fmt.Fprintln(w, `{"error": "Need to specify from and to URL parameters"}`)
		return
	}
	_, _, _ = from, to, rpc

	// TODO:  send RPC call, and respond on HTTP request
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintln(w, `{ piechart: { labels: ["Schoolwork","Netflix and chill"], data: [300,50] }, last: true, first: true, total: 350 }`)
}
