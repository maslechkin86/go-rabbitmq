package main

import (
	"context"
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

	if _, err := client.CreateQueue("customers_created", true, false); err != nil {
		panic(err)
	}
	if _, err := client.CreateQueue("customers_test", false, true); err != nil {
		panic(err)
	}

	if err := client.CreateBinding("customers_created", "customers.created.*", "customer_events"); err != nil {
		panic(err)
	}
	if err := client.CreateBinding("customers_test", "customers.*", "customer_events"); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Send(ctx, "customer_events", "customers.created.se", amqp.Publishing{
		ContentType:  "text/plain",
		DeliveryMode: amqp.Persistent, // This tells rabbitMQ that this message should be Saved if no resources accepts it before a restart (durable)
		Body:         []byte("An cool message between services"),
	}); err != nil {
		panic(err)
	}
	if err := client.Send(ctx, "customer_events", "customers.test", amqp.Publishing{
		ContentType:  "text/plain",
		DeliveryMode: amqp.Transient, // This tells rabbitMQ that this message can be deleted if no resources accepts it before a restart (non durable)
		Body:         []byte("A second cool message"),
	}); err != nil {
		panic(err)
	}

	log.Println(client)
}
