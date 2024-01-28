package main

import (
	"context"
	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
	"log"
	"programmingpercy.tech/eventdrivenrabbit/internal"
	"time"
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

	var blocking chan struct{}

	ConsumeMessagesConcurrently(messageBus)

	log.Println("Consuming, to close the program press CTRL+C")

	// This will block forever
	<-blocking
}

// ConsumeMessages consume messages from channel.
func ConsumeMessages(messageBus <-chan amqp091.Delivery) {
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
}

// ConsumeMessagesConcurrently consume messages from channel using a work group that allows for a certain amount of concurrent tasks.
func ConsumeMessagesConcurrently(messageBus <-chan amqp091.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tasksCount := 10
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(tasksCount)
	go func() {
		for message := range messageBus {
			msg := message
			g.Go(func() error {
				log.Printf("New Message: %v", msg)

				time.Sleep(10 * time.Second)
				if err := msg.Ack(false); err != nil {
					log.Printf("Acknowledged message failed: Retry ? Handle manually %s\n", msg.MessageId)
					return err
				}
				log.Printf("Acknowledged message %s\n", msg.MessageId)
				return nil
			})
		}
	}()
}
