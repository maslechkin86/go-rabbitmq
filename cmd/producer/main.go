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

	// CreateQueuesAndBindings(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		if err := client.SendAndWait(ctx, "customer_events", "customers.created.us", amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:         []byte(`A ne message`),
		}); err != nil {
			panic(err)
		}
	}

	log.Println(client)
}

func CreateQueuesAndBindings(client internal.RabbitClient) {
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

}
