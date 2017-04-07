package main

import (
	"github.com/pajjme/iTime-client-api/api"
	"github.com/streadway/amqp"
	"net/http"
	"log"
)

const ApiVersion = "/v1"
const AmqpUrl = "amqp://guest:guest@localhost:5672/"
const HttpAddr = ":8118"

func main() {
	log.Print("Starting HTTP server")

	log.Printf("Connecting to '%s'", AmqpUrl)
	conn, err := amqp.Dial(AmqpUrl)
	api.CheckError(err)
	defer conn.Close()

	channel, err := conn.Channel()
	api.CheckError(err)
	defer channel.Close()

	qm := api.MakeAmqpRPC(*channel)

	// Bind each handler to channel and an endpoint
	mux := http.NewServeMux()

	// TODO: Make sure AMQP-connection works, ex reconnect
	mux.HandleFunc(ApiVersion + "/authorize", func(w http.ResponseWriter, r *http.Request) {
		api.Authorize(w, r, qm)
	})
	mux.HandleFunc(ApiVersion + "/stats", func(w http.ResponseWriter, r *http.Request) {
		api.Stats(w, r, qm)
	})

	log.Printf("Start listening for HTTP connections at '%s'", HttpAddr)
	err = http.ListenAndServe(HttpAddr, mux)
	api.CheckError(err)
}
