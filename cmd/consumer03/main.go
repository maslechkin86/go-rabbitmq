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

	publishConn, err := internal.ConnectRabbitMQ("guest", "guest", "localhost:5672", "customers")
	if err != nil {
		panic(err)
	}
	defer func() { _ = publishConn.Close() }()

	publishClient, err := internal.NewRabbitMQClient(publishConn)
	if err != nil {
		panic(err)
	}
	defer func() { _ = publishClient.Close() }()

	queue, err := client.CreateQueue("", true, true)
	if err != nil {
		panic(err)
	}

	if err = client.CreateBinding(queue.Name, "", "customer_events"); err != nil {
		panic(err)
	}

	messageBus, err := client.Consume(queue.Name, "email-service", false)
	if err != nil {
		panic(err)
	}

	// Set a timeout for 15 secs
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	maxMessagesCount := 10
	maxBytesSize := 0
	if err := client.ApplyQos(maxMessagesCount, maxBytesSize, true); err != nil {
		panic(err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	go func() {
		for message := range messageBus {
			msg := message
			g.Go(func() error {
				log.Printf("New message: %v\n", message)
				time.Sleep(10 * time.Second)
				if err := msg.Ack(false); err != nil {
					log.Println("Acknowledge message failed")
					return err
				}
				if err := publishClient.SendAndWait(ctx, "customer_callbacks", msg.ReplyTo, amqp091.Publishing{
					ContentType:   "text/plain",
					DeliveryMode:  amqp091.Persistent,
					Body:          []byte("RPC Complete"),
					CorrelationId: msg.CorrelationId,
				}); err != nil {
					panic(err)
				}
				log.Printf("Acknowledge message %s\n", message.MessageId)
				return nil
			})
		}
	}()

	// This will block forever
	var blocking chan struct{}
	log.Println("Consuming, to close the program press CTRL+C")
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
