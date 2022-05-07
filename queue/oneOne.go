package queue

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
)

type oneOne struct {
	queueName string
	queue     *queue
}

func (q *queue) NewOneOne(queueName string) *oneOne {
	return &oneOne{
		queueName: queueName,
		queue:     q,
	}
}

func (o *oneOne) execute(f func(*amqp.Channel) error) error {
	execFunc := func(ch *amqp.Channel) error {
		// With the instance and declare Queues that we can
		// publish and subscribe to.
		if _, err := ch.QueueDeclare(
			o.queueName, // queue name
			true,        // durable
			false,       // auto delete
			false,       // exclusive
			false,       // no wait
			nil,         // arguments
		); err != nil {
			panic(err)
		}
		return f(ch)
	}

	return o.queue.execute(execFunc)
}

func (o *oneOne) Send(info interface{}) error {

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
			"",          // exchange
			o.queueName, // queue name
			false,       // mandatory
			false,       // immediate
			message,     // message to publish
		); err != nil {
			return err
		}
		return nil
	}
	return o.execute(sendFunc)
}

func (o *oneOne) Subscribe(task func(messageFromQueue Message)) error {
	subscribeFunc := func(channel *amqp.Channel) error {
		// Subscribing to QueueService1 for getting messages.
		messages, err := channel.Consume(
			o.queueName, // queue name
			"",          // consumer
			true,        // auto-ack
			false,       // exclusive
			false,       // no local
			false,       // no wait
			nil,         // arguments
		)
		if err != nil {
			return err
		}

		// Build a welcome message.
		log.Printf("Successfully connected to RabbitMQ on \"%s\" queue", o.queueName)
		log.Println("Waiting for messages")

		// Make a channel to receive messages into infinite loop.
		forever := make(chan bool)

		go func() {
			for message := range messages {
				// show received message in a console.
				log.Printf(" > Received message: %s\n", message.Body)
				// Do the task
				task(Message{message})
			}
		}()

		<-forever

		return nil
	}
	return o.execute(subscribeFunc)
}
