package main

import (
	"pajjme/client/api"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Substitute for UUID. Will be replaced.
func randomString(length int) string {
	randInt := func(min int, max int) int {
		return min + rand.Intn(max-min)
	}

	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

// Structure for kepping track of requests and responses to AMQP.
type QueueManager struct {
	// TODO: Requests that doesn't get a response fill up getexpectedResponses
	expectedResponses map[string](chan []byte)
	amqpChannel       amqp.Channel
	amqpQueue         amqp.Queue
}

// Makes a QueueManager from a AMQP manager.
// Starts a goroutine that redirects responses to corresponding
// gochannel so it can be used by the goroutines
func makeQueueManager(amqpChannel amqp.Channel) (qm QueueManager) {
	// TODO: better to create queue in init function?
	queue, err := amqpChannel.QueueDeclare("", false, true, true, false, nil)
	checkError(err)

	qm = QueueManager{make(map[string]chan []byte), amqpChannel, queue}

	// Wait for incoming AMQP messages, then forward the body to the requester
	consumer, err := amqpChannel.Consume(queue.Name, "", false, true, false, false, nil)
	checkError(err)
	go func() {
		for msg := range consumer {
			answered, ok := qm.expectedResponses[msg.CorrelationId]

			if !ok {
				// TODO: use logger instead
				println("Warning: Got response without expecting correlation ID '" + msg.CorrelationId + "'. ")
				continue
			}

			answered <- msg.Body
			delete(qm.expectedResponses, msg.CorrelationId)
		}
	}()
	return
}

// Adds an entry in the QueueManager for a request and returns the gochannel.
func (qm QueueManager) sendRequest(endpoint string, body []byte) chan []byte {
	//ish uuid
	corrId := randomString(32)
	respondChannel := make(chan []byte)

	qm.expectedResponses[corrId] = respondChannel

	log.Println("endpoint " + endpoint)
	err := qm.amqpChannel.Publish("", endpoint, false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       qm.amqpQueue.Name,
			Body:          body,
		},
	)
	if err != nil {
		delete(qm.expectedResponses, qm.amqpQueue.Name)
		checkError(err)
	}
	return respondChannel
}

const ApiVersion = "/v1"
const AmqpUrl = "amqp://guest:guest@localhost:5672/"
const HttpAddr = ":8118"

func main() {

	log.Println("Starting the server")

	var err error
	var conn *amqp.Connection

	for {
		conn, err = amqp.Dial(os.Getenv("AMQP_URL"))
		if conn != nil {
			break
		}
		log.Println("Cannot connect to rabbitmq. Retrying... ")
		time.Sleep(5 * time.Second)

	}
	log.Println("Connected to rabbitmq")
	defer conn.Close()

	channel, err := conn.Channel()
	api.CheckError(err)
	defer channel.Close()

	qm := api.MakeAmqpRPC(*channel)

	// Bind each handler to channel and an endpoint
	mux := http.NewServeMux()

	// TODO: Make sure AMQP-connection works, ex reconnect
	mux.HandleFunc(ApiVersion+"/authorize", func(w http.ResponseWriter, r *http.Request) {
		api.Authorize(w, r, qm)
	})
	mux.HandleFunc(ApiVersion+"/stats", func(w http.ResponseWriter, r *http.Request) {
		api.Stats(w, r, qm)
	})

	log.Printf("Start listening for HTTP connections at '%s'", HttpAddr)
	err = http.ListenAndServe(HttpAddr, mux)
	api.CheckError(err)
}
