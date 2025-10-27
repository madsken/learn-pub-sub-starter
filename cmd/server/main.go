package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	pubCh, err := amqpConnection.Channel()
	if err != nil {
		log.Fatalf("could not create publish channel: %s", err)
	}

	err = createSubscribers(amqpConnection)
	if err != nil {
		log.Fatalf("could not create subscriptions: %s", err)
	}

	gamelogic.PrintServerHelp()

	for {
		commands := gamelogic.GetInput()
		if len(commands) == 0 {
			continue
		}

		switch commands[0] {
		case "pause":
			fmt.Println("Publishing pause: true")
			err = pubsub.PublishJSON(pubCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Printf("Could not publish in time: %v\n", err)
			}
		case "resume":
			fmt.Println("Publishing pause: false")
			pubsub.PublishJSON(pubCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Printf("Could not publish in time: %v\n", err)
			}
		case "quit":
			fmt.Println("Closing server...")
			return
		default:
			fmt.Println("Invalid command")
		}
	}
}

func createSubscribers(conn *amqp.Connection) error {
	err := pubsub.SubscribeGob(
		conn,
		string(routing.ExchangePerilTopic),
		string(routing.GameLogSlug),
		string(routing.GameLogSlug)+".*",
		pubsub.Durable,
		handlerLogs(),
	)
	if err != nil {
		return err
	}

	return nil
}
