package main

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"github.com/streadway/amqp"
)

const ApiVersion string = "/v1"
const AmqpUrl  = "amqp://guest:guest@localhost:5672/"
const UrlMapping  = map[string] func(http.ResponseWriter, *http.Request){
	"/authorize": authorize,
	"/stats": stats,
}

type authorizeRequest struct {
	SessionToken string `json:"session_token"`
}

func checkError(err error)  {
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	print("starting server")

	conn, err := amqp.Dial(AmqpUrl)
	checkError(err)
	defer conn.Close()

	channel, err := conn.Channel()
	checkError(err)
	defer channel.Close()


	mux := http.NewServeMux()
	for url, handler := range UrlMapping {
		queue,err := channel.QueueDeclare(
			ApiVersion + url,
			false,
			false,
			false,
			false,
			nil
		)
		// TODO: How do we give a queue to each handler
		checkError(err)
		mux.HandleFunc(ApiVersion + url, handler)
	}
	err = http.ListenAndServe(":8118", mux)
	checkError(err)
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