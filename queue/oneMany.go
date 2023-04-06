package queue

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type oneMany struct {
	exchangeName string
	queue        *queue
}

func (q *queue) NewOneMany(exchangeName string) *oneMany {
	return &oneMany{
		exchangeName: exchangeName,
		queue:        q,
	}
}

func (o *oneMany) execute(f func(*amqp.Channel) error) error {
	execFunc := func(ch *amqp.Channel) error {
		err := ch.ExchangeDeclare(
			o.exchangeName, // name
			"fanout",       // type
			true,           // durable
			false,          // auto-deleted
			false,          // internal
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			panic(err)
		}
		return f(ch)
	}

	return o.queue.execute(execFunc)
}

func (o *oneMany) Send(info interface{}) error {

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
			o.exchangeName, // exchange
			"",             // queue name
			false,          // mandatory
			false,          // immediate
			message,        // message to publish
		); err != nil {
			return err
		}
		return nil
	}
	return o.execute(sendFunc)
}

func (o *oneMany) Subscribe(task func(messageFromQueue Message)) error {
	subscribeFunc := func(channel *amqp.Channel) error {
		q, err := channel.QueueDeclare(
			"",    // name
			true,  // durable
			true,  // auto delete
			false, // exclusive
			false, // no wait
			nil,   // arguments
		)
		if err != nil {
			return err
		}

		err = channel.QueueBind(
			q.Name,         // queue name
			"",             // routing key
			o.exchangeName, // exchange
			false,
			nil,
		)
		// Subscribing to QueueService1 for getting messages.
		messages, err := channel.Consume(
			q.Name, // queue name
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no local
			false,  // no wait
			nil,    // arguments
		)
		if err != nil {
			return err
		}

		// Build a welcome message.
		log.Printf("Successfully connected to RabbitMQ on \"%s\" queue and \"%s\" exchange", q.Name, o.exchangeName)
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
