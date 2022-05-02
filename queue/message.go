package queue

import "github.com/streadway/amqp"

type Message struct {
	amqp.Delivery
}
