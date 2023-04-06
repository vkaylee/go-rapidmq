package queue

import amqp "github.com/rabbitmq/amqp091-go"

type Message struct {
	amqp.Delivery
}
