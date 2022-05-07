package queue

import "github.com/streadway/amqp"

type queue struct {
	// Define RabbitMQ server URL.
	amqpServerAddress string
}

func NewQueue(amqpServerAddress string) *queue {
	return &queue{
		amqpServerAddress: amqpServerAddress,
	}
}

func (q *queue) execute(f func(*amqp.Channel) error) error {
	// Create a new RabbitMQ connection.
	conn, err := amqp.Dial(q.amqpServerAddress)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Let's start by opening a channel to our RabbitMQ
	// instance over the connection we have already
	// established.
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	return f(ch)
}
