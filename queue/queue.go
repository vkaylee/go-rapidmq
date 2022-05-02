package queue

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
)

type queue struct {
	// Define RabbitMQ server URL.
	amqpServerAddress string
	amqpChannelName   string
}

func NewQueue(amqpServerAddress string, amqpChannelName string) *queue {
	return &queue{
		amqpServerAddress: amqpServerAddress,
		amqpChannelName:   amqpChannelName,
	}
}

func (q *queue) execute(f func(*amqp.Channel) error) error {
	// Create a new RabbitMQ connection.
	connectRabbitMQ, err := amqp.Dial(q.amqpChannelName)
	if err != nil {
		panic(err)
	}
	defer connectRabbitMQ.Close()

	// Let's start by opening a channel to our RabbitMQ
	// instance over the connection we have already
	// established.
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}
	defer channelRabbitMQ.Close()
	// With the instance and declare Queues that we can
	// publish and subscribe to.
	if _, err := channelRabbitMQ.QueueDeclare(
		q.amqpChannelName, // queue name
		true,              // durable
		false,             // auto delete
		false,             // exclusive
		false,             // no wait
		nil,               // arguments
	); err != nil {
		panic(err)
	}
	return f(channelRabbitMQ)
}

func (q *queue) Send(info interface{}) error {

	sendFunc := func(channel *amqp.Channel) error {
		// Encode to byte
		data, err := json.Marshal(info)
		if err != nil {
			return err
		}
		// Create a message to publish.
		message := amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		}

		// Attempt to publish a message to the queue.
		if err := channel.Publish(
			"",                // exchange
			q.amqpChannelName, // queue name
			false,             // mandatory
			false,             // immediate
			message,           // message to publish
		); err != nil {
			return err
		}
		return nil
	}
	return q.execute(sendFunc)
}

func (q *queue) Subscribe(task func(messageFromQueue amqp.Delivery)) error {
	subscribeFunc := func(channel *amqp.Channel) error {
		// Subscribing to QueueService1 for getting messages.
		messages, err := channel.Consume(
			q.amqpChannelName, // queue name
			"",                // consumer
			true,              // auto-ack
			false,             // exclusive
			false,             // no local
			false,             // no wait
			nil,               // arguments
		)
		if err != nil {
			log.Println(err)
		}

		// Build a welcome message.
		log.Printf("Successfully connected to RabbitMQ, ")
		log.Println("Waiting for messages")

		// Make a channel to receive messages into infinite loop.
		forever := make(chan bool)

		go func() {
			for message := range messages {
				// show received message in a console.
				log.Printf(" > Received message: %s\n", message.Body)
				// Do the task
				task(message)
			}
		}()

		<-forever

		return nil
	}
	return q.execute(subscribeFunc)
}
