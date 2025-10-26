package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const connectionString = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril client...")
	amqpConnection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Error creating RabbitMQ connection: %s", err)
	}
	defer amqpConnection.Close()

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error creating client: %s", err)
	}

	pubsub.DeclareAndBind(
		amqpConnection,
		"peril_direct",
		fmt.Sprintf("%s.%s", routing.PauseKey, userName),
		routing.PauseKey,
		pubsub.Transient,
	)

	fmt.Println("Peril client started successfully...")

	waitForSigint()

	fmt.Println("Closing client...")
}

func waitForSigint() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
