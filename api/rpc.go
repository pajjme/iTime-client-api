package api

import (
	"log"
	"github.com/streadway/amqp"
)

type RPCaller interface {
	SendRequest(method string, body []byte) chan []byte
}

// Structure for kepping track of requests and responses to AMQP.
type AmqpRPC struct {
	// TODO: Requests that doesn't get a response fill up getexpectedResponses
	expectedResponses map[string](chan []byte)
	amqpChannel       amqp.Channel
	amqpQueue         amqp.Queue
}

// Makes a QueueManager from a AMQP manager.
// Starts a goroutine that redirects responses to corresponding
// gochannel so it can be used by the goroutines
func MakeAmqpRPC(amqpChannel amqp.Channel) (qm AmqpRPC) {
	// TODO: better to create queue in init function?
	queue, err := amqpChannel.QueueDeclare("", false, true, true, false, nil)
	CheckError(err)
	qm = AmqpRPC{make(map[string]chan []byte), amqpChannel, queue}

	// Wait for incoming AMQP messages, then forward the body to the requester
	consumer, err := amqpChannel.Consume(queue.Name, "", false, true, false, false, nil)
	CheckError(err)
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
func (qm AmqpRPC) SendRequest(method string, body []byte) chan []byte {
	//ish uuid
	corrId := RandomString(32)
	respondChannel := make(chan []byte)

	qm.expectedResponses[corrId] = respondChannel

	log.Println("endpoint " + method)
	err := qm.amqpChannel.Publish("", method, false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       qm.amqpQueue.Name,
			Body:          body,
		},
	)
	if err != nil {
		delete(qm.expectedResponses, qm.amqpQueue.Name)
		CheckError(err)
	}
	return respondChannel
}
