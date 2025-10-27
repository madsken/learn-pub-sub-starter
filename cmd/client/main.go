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

	amqpChan, amqpQueue, err := pubsub.DeclareAndBind(
		amqpConnection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, userName),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatalf("error subscribing: %s", err)
	}
	fmt.Printf("successfully bound to: %s\n", amqpQueue.Name)

	gameState := gamelogic.NewGameState(userName)
	err = createSubscribers(gameState, amqpConnection, amqpChan)
	if err != nil {
		log.Fatalf("could not create subscriptions: %s", err)
	}

	for {
		commands := gamelogic.GetInput()
		if len(commands) == 0 {
			continue
		}

		switch commands[0] {
		case "spawn":
			err = gameState.CommandSpawn(commands)
			if err != nil {
				fmt.Printf("invalid subcommand: %v\n", err)
				continue
			}
		case "move":
			move, err := gameState.CommandMove(commands)
			if err != nil {
				fmt.Printf("invalid subcommand: %v\n", err)
				continue
			}

			err = pubsub.PublishJSON(amqpChan, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+gameState.GetUsername(), move)
			if err != nil {
				fmt.Printf("could not publish move: %s\n", err)
				continue
			}
			fmt.Println("Move published to players")

		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("No spamming allowed")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Invalid command")
		}
	}
}

func createSubscribers(gs *gamelogic.GameState, conn *amqp.Connection, amqpChan *amqp.Channel) error {
	err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	)
	if err != nil {
		return err
	}

	err = pubsub.SubscribeJSON(
		conn,
		string(routing.ExchangePerilTopic),
		string(routing.ArmyMovesPrefix)+"."+gs.GetUsername(),
		string(routing.ArmyMovesPrefix)+".*",
		pubsub.Transient,
		handlerMove(gs, amqpChan),
	)
	if err != nil {
		return err
	}

	err = pubsub.SubscribeJSON(
		conn,
		string(routing.ExchangePerilTopic),
		string(routing.WarRecognitionsPrefix),
		string(routing.WarRecognitionsPrefix)+".*",
		pubsub.Durable,
		handlerWar(gs, amqpChan),
	)
	if err != nil {
		return err
	}

	return nil
}
