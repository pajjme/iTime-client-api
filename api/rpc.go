package api

import (
	"log"
	"github.com/streadway/amqp"
)

type RPCaller interface {
	SendRequest(method string, body []byte) chan []byte
}

type AmqpRPC struct {
	// TODO: Requests that doesn't get a response fill up getexpectedResponses
	expectedResponses map[string](chan []byte)
	amqpChannel       amqp.Channel
	amqpQueue         amqp.Queue
}

// Starts a go routine that redirects AMQP responses to corresponding
// go channel in expectedResponses
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
				log.Printf("Warning: Got response without expecting correlation ID '%s'.", msg.CorrelationId)
				continue
			}

			answered <- msg.Body
			delete(qm.expectedResponses, msg.CorrelationId)
		}
	}()
	return
}

func (qm AmqpRPC) SendRequest(method string, body []byte) chan []byte {
	corrId := RandomString(32)
	respondChannel := make(chan []byte)
	qm.expectedResponses[corrId] = respondChannel

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
