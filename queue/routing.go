package queue

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type routing struct {
	exchangeName string
	queue        *queue
}

func (q *queue) NewRouting(exchangeName string) *routing {
	return &routing{
		exchangeName: exchangeName,
		queue:        q,
	}
}

func (o *routing) execute(f func(*amqp.Channel, context.Context) error) error {
	execFunc := func(ch *amqp.Channel, ctx context.Context) error {
		err := ch.ExchangeDeclare(
			o.exchangeName, // name
			"direct",       // type
			true,           // durable
			false,          // auto-deleted
			false,          // internal
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			panic(err)
		}
		return f(ch, ctx)
	}

	return o.queue.execute(execFunc)
}

func (o *routing) Send(info interface{}, routingKey string) error {

	sendFunc := func(channel *amqp.Channel, ctx context.Context) error {
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
		if err := channel.PublishWithContext(
			ctx,
			o.exchangeName, // exchange
			routingKey,     // queue name, routing key
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

func (o *routing) Subscribe(task func(messageFromQueue Message), taskTypes []string) error {
	subscribeFunc := func(channel *amqp.Channel, ctx context.Context) error {
		q, err := channel.QueueDeclare(
			"",    // name
			false, // durable
			false, // delete when unused
			true,  // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return err
		}
		for _, taskType := range taskTypes {
			err = channel.QueueBind(
				q.Name,         // queue name
				taskType,       // routing key
				o.exchangeName, // exchange
				false,
				nil,
			)
		}
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
