package main

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	internal "programmingpercy.tech/eventdrivenrabbit/internal"
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

	consumeConn, err := internal.ConnectRabbitMQ("guest", "guest", "localhost:5672", "customers")
	if err != nil {
		panic(err)
	}
	defer func() { _ = consumeConn.Close() }()

	consumeClient, err := internal.NewRabbitMQClient(consumeConn)
	if err != nil {
		panic(err)
	}
	defer func() { _ = consumeClient.Close() }()

	queue, err := consumeClient.CreateQueue("", true, true)
	if err != nil {
		panic(err)
	}

	if err = consumeClient.CreateBinding(queue.Name, queue.Name, "customer_callbacks"); err != nil {
		panic(err)
	}

	messageBus, err := consumeClient.Consume(queue.Name, "customer-api", true)

	go func() {
		for message := range messageBus {
			log.Printf(" Message Callback %s\n", message.CorrelationId)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		if err := client.SendAndWait(ctx, "customer_events", "customers.created.us", amqp.Publishing{
			ContentType:   "text/plain",
			DeliveryMode:  amqp.Persistent,
			ReplyTo:       queue.Name,
			CorrelationId: fmt.Sprintf("customer_created_%d", i),
			Body:          []byte(`A ne message`),
		}); err != nil {
			panic(err)
		}
	}

	// This will block forever
	var blocking chan struct{}
	log.Println("Consuming, to close the program press CTRL+C")
	<-blocking
}
