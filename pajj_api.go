package main

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"github.com/streadway/amqp"
	"time"
)

const ApiVersion string = "/v1"
const AmqpUrl = "amqp://guest:guest@localhost:5672/"
const UrlMapping = map[string]func(http.ResponseWriter, *http.Request, *QueueManager){
	"/authorize": authorize,
	"/stats": stats,
}

type authorizeRequest struct {
	SessionToken string `json:"session_token"`
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max - min)
}

type QueueManager struct {
	// TODO: Requests that doesn't get a response fill up getexpectedResponses
	expectedResponses map[string](<-chan amqp.Delivery)
	amqpChannel       amqp.Channel
	amqpQueue         amqp.Queue
}

func makeQueueManager(amqpChannel amqp.Channel) (qm QueueManager) {
	// TODO: better to create queue in init function?
	queue, err := amqpChannel.QueueDeclare("", false, true, true, false, nil)
	checkError(err)

	qm = QueueManager{make(map[string]struct{}), amqpChannel, queue}

	// Wait for incoming AMQP messages, then forward the body to the requester
	consumer, err := amqpChannel.Consume(queue.Name, "", false, true, false, false, nil)
	checkError(err)
	go func() {
		for msg := range consumer {
			answered, ok := qm.expectedResponses[msg.CorrelationId]
			defer delete(qm.expectedResponses, msg.CorrelationId)

			if !ok {
				// TODO: use logger instead
				println("Warning: Got response without expecting correlation ID '" + msg.CorrelationId + "'. ")
				continue
			}

			answered <- msg.Body
		}
	}()
}

func (qm QueueManager) sendRequest(endpoint string, body string) <-chan amqp.Delivery {
	corrId := randomString(32)
	respondChannel := make(<-chan amqp.Delivery)

	qm.expectedResponses[qm.amqpQueue.Name] = respondChannel

	err := qm.amqpChannel.Publish("", endpoint, true, true, amqp.Publishing{
		ContentType: "application/json",
		CorrelationId: corrId,
		ReplyTo: qm.amqpQueue.Name,
		Body: body,
	})
	if err {
		delete(qm.expectedResponses, qm.amqpQueue.Name)
		checkError(err)
	}
}

func callRPC(endpoint string, channel amqp.Channel) []byte {
	cons, err := channel.Consume(endpoint, "", true, false, false, false, nil)
	checkError(err)

	err := channel.Publish("", endpoint)
}

func main() {
	print("starting server")

	conn, err := amqp.Dial(AmqpUrl)
	checkError(err)
	defer conn.Close()

	channel, err := conn.Channel()
	checkError(err)
	defer channel.Close()

	qm := makeQueueManager(channel)

	// Bind each handler to channel and an endpoint
	mux := http.NewServeMux()
	for url, handler := range UrlMapping {
		mux.HandleFunc(ApiVersion + url, func(w http.ResponseWriter, r *http.Request) {
			// TODO: Make sure AMQP-connection works, ex reconnect
			handler(w, r, &qm)
		})
	}

	err = http.ListenAndServe(":8118", mux)
	checkError(err)
}

func authorize(w http.ResponseWriter, r *http.Request, qm *QueueManager) {
	authReq := authorizeRequest{}
	jsonText, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(jsonText, &authReq)
	checkError(err)


	amqpResponse := <- qm.sendRequest("authorize", json.Marshal(jsonText))

	// TODO: Use data from amqpResponse to send to client

	w.WriteHeader(200)

	http.SetCookie(w, http.Cookie{
		Name: "sessionToken",
		Value:,
		Expires: time.Now().AddDate(1, 0, 0), // One year ahead
	})

	fmt.Fprintln(w, "{}")
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