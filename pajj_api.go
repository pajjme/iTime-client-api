package main

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const ApiVersion string = "/v1"
const AmqpUrl = "amqp://guest:guest@localhost:5672/"

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

func main() {
	
	log.Println("Starting the server")
	conn, err := amqp.Dial(AmqpUrl)
	checkError(err)
	defer conn.Close()
	channel, err := conn.Channel()
	checkError(err)
	defer channel.Close()

	qm := makeQueueManager(*channel)

	// Bind each handler to channel and an endpoint
	mux := http.NewServeMux()

	// TODO: Make sure AMQP-connection works, ex reconnect
	mux.HandleFunc(ApiVersion+"/authorize",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			authorize(w, r, &qm)
		},
	)
	mux.HandleFunc(ApiVersion+"/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		stats(w, r, &qm)
	})

	err = http.ListenAndServe(":8118", mux)
	checkError(err)
}

type authorizeRequest struct {
	// Makes it possible to Marshal the struct to json.
	AuthCode string `json:"auth_code"`
}

func authorize(w http.ResponseWriter, r *http.Request, qm *QueueManager) {
	authReq := authorizeRequest{}
	jsonText, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(jsonText, &authReq)
	checkError(err)


	
	log.Println("Send request to US", authReq)
	rpcRequest, err := json.Marshal(authReq)
	amqpResponse := <-qm.sendRequest("authorize", rpcRequest)
	log.Println("Response from US was:",string(amqpResponse))
	// TODO: Use data from amqpResponse to send to client

	//HTTP Response: Found
	w.WriteHeader(200)

	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   "",
		Expires: time.Now().AddDate(1, 0, 0), // One year ahead
	})

	fmt.Fprintln(w, string(amqpResponse))
}

func stats(w http.ResponseWriter, r *http.Request, qm *QueueManager) {
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
	fmt.Fprintln(w, "{}")
}
