package main

import (
	"log"
	"programmingpercy.tech/eventdrivenrabbit/internal"
)

func main() {
	conn, err := internal.ConnectRabbitMQ("guest", "guest", "localhost:5672", "customers")
	if err != nil {
		panic(err)
	}
	defer func() { _ = conn.Close() }()

	client, err := internal.NewRabbitMQClient(conn)
	if err != nil {
		panic(err)
	}
	defer func() { _ = client.Close() }()

	messageBus, err := client.Consume("customers_created", "email-service", false)
	if err != nil {
		panic(err)
	}

	go func() {
		for message := range messageBus {
			log.Printf("New message: %v\n", message)
			if !message.Redelivered {
				message.Nack(false, true)
				continue
			}
			if err := message.Ack(false); err != nil {
				log.Println("Acknowledge message failed")
				continue
			}
			log.Printf("Acknowledge message %s\n", message.MessageId)
		}
	}()

	// This will block forever
	var blocking chan struct{}
	log.Println("Consuming, to close the program press CTRL+C")
	<-blocking
}
