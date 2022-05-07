# go-rapidmq

## Install

```shell
  go get github.com/vleedev/go-rapidmq:v0.2.0
```
## Simple code to create job queue

One message can be received by one subscriber
```go
package main

import (
	"github.com/vleedev/go-rapidmq/queue"
	"log"
	"os"
)

func main() {
	role := os.Args[1]
	queueServer := queue.NewQueue("amqp://guest:guest@127.0.0.1:5672/")
	myQueue := queueServer.NewOneOne("queueName1")
	if role == "sender" {
		message := os.Args[2]
		if err := myQueue.Send(message); err != nil {
			log.Fatalln(err)
		}
	}
	
	if role == "subscriber" {
		task := func(message queue.Message) {
			log.Printf("Message from sender: %s\n", message.Body)
		}
		if err := myQueue.Subscribe(task); err != nil {
			log.Fatalln(err)
		}
	}
}
```
- Run `go run main.go subscriber` to subscribe the queue.
- Run `go run main.go sender "Test message"` to send the test message to the queue

## Simple code to create broadcast queue

One message can be received by all subscribers
```go
package main

import (
	"github.com/vleedev/go-rapidmq/queue"
	"log"
	"os"
)

func main() {
	role := os.Args[1]
	queueServer := queue.NewQueue("amqp://guest:guest@127.0.0.1:5672/")
	myQueue := queueServer.NewOneMany("exchangeName1")
	if role == "sender" {
		message := os.Args[2]
		if err := myQueue.Send(message); err != nil {
			log.Fatalln(err)
		}
	}
	
	if role == "subscriber" {
		task := func(message queue.Message) {
			log.Printf("Message from sender: %s\n", message.Body)
		}
		if err := myQueue.Subscribe(task); err != nil {
			log.Fatalln(err)
		}
	}
}
```
- Run `go run main.go subscriber` to subscribe the queue.
- Run `go run main.go sender "Test message"` to send the test message to the queue