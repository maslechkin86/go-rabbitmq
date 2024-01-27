package main

import (
	"context"
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
	defer conn.Close()
	client, err := internal.NewRabbitMQClient(conn)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	messageBus, err := client.Consume("customers_created", "email-service", false)
	if err != nil {
		panic(err)
	}

	var blocking chan struct{}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	// Set amount of concurrent tasks
	g.SetLimit(10)
	go func() {
		for message := range messageBus {
			// Spawn a worker
			msg := message
			g.Go(func() error {
				log.Printf("New Message: %v", msg)

				time.Sleep(10 * time.Second)
				// Multiple means that we acknowledge a batch of messages, leave false for now
				if err := msg.Ack(false); err != nil {
					log.Printf("Acknowledged message failed: Retry ? Handle manually %s\n", msg.MessageId)
					return err
				}
				log.Printf("Acknowledged message %s\n", msg.MessageId)
				return nil
			})
		}
	}()
	//var blocking chan struct{}
	//go func() {
	//	for message := range messageBus {
	//		log.Printf("New message: %v\n", message)
	//		if !message.Redelivered {
	//			message.Nack(false, true)
	//			continue
	//		}
	//		if err := message.Ack(false); err != nil {
	//			log.Println("Acknowledge message failed")
	//			continue
	//		}
	//		log.Printf("Acknowledge message %s\n", message.MessageId)
	//	}
	//}()

	log.Println("Consuming, to close the program press CTRL+C")
	// This will block forever
	<-blocking
}
