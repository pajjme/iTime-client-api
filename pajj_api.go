package main

import (
	"github.com/pajjme/iTime-client-api/apiutil"
	"github.com/streadway/amqp"
	"net/http"
)

const ApiVersion = "/v1"
const AmqpUrl = "amqp://guest:guest@localhost:5672/"

func main() {
	print("starting server")

	conn, err := amqp.Dial(AmqpUrl)
	apiutil.CheckError(err)
	defer conn.Close()

	channel, err := conn.Channel()
	apiutil.CheckError(err)
	defer channel.Close()

	qm := apiutil.MakeAmqpRPC(*channel)

	// Bind each handler to channel and an endpoint
	mux := http.NewServeMux()

	// TODO: Make sure AMQP-connection works, ex reconnect
	mux.HandleFunc(ApiVersion+"/authorize", func(w http.ResponseWriter, r *http.Request) {
		apiutil.Authorize(w, r, qm)
	})
	mux.HandleFunc(ApiVersion+"/stats", func(w http.ResponseWriter, r *http.Request) {
		apiutil.Stats(w, r, qm)
	})

	err = http.ListenAndServe(":8118", mux)
	apiutil.CheckError(err)
}
