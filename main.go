package main

import (
	"github.com/vleedev/go-rapidmq/queue"
	"log"
	"os"
)

func main() {
	firstArgv := os.Args[1]
	secondArgv := os.Args[2]
	//amqpServerURL := os.Getenv("AMQP_SERVER_URL")
	myqueue := queue.NewQueue("amqp://guest:guest@127.0.0.1:5672/")
	if firstArgv == "oneOne" {
		queueName := os.Args[3]
		oneOne := myqueue.NewOneOne(queueName)
		if secondArgv == "sender" {
			oneOne.Send("hello")
		} else if secondArgv == "subscriber" {
			task := func(message queue.Message) {
				log.Printf("Xin chao message: %s\n", message.Body)
			}
			err := oneOne.Subscribe(task)
			if err != nil {
				log.Fatalln(err)
			}
		}

	} else if firstArgv == "oneMany" {
		oneMany := myqueue.NewOneMany("myexchange")
		if secondArgv == "sender" {
			i := 1
			for true {
				oneMany.Send(i)
				i++
				if i == 100000 {
					break
				}
			}
		} else if secondArgv == "subscriber" {
			task := func(message queue.Message) {
				log.Printf("Xin chao message: %s\n", message.Body)
			}
			err := oneMany.Subscribe(task)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}
