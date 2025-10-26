package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const connectionString = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")
	amqpConnection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Error creating RabbitMQ connection: %s", err)
	}
	defer amqpConnection.Close()

	amqpChan, err := amqpConnection.Channel()
	if err != nil {
		log.Fatalf("Could not create RabbitMQ channel: %s", err)
	}

	pubsub.PublishJSON(amqpChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: false,
	})

	fmt.Println("Peril server started successfully...")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Closing server...")
}
