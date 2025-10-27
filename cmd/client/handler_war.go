package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerWar(gs *gamelogic.GameState, amqpChan *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(row gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(row)
		gl := routing.GameLog{
			CurrentTime: time.Now(),
			Username:    gs.GetUsername(),
		}

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeYouWon:
			gl.Message = fmt.Sprintf("%s won a war against %s\n", winner, loser)
			err := publishGameLog(gl, amqpChan)
			if err != nil {
				fmt.Printf("error publishing war message: %s", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			gl.Message = fmt.Sprintf("%s won a war against %s\n", winner, loser)
			err := publishGameLog(gl, amqpChan)
			if err != nil {
				fmt.Printf("error publishing war message: %s", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			gl.Message = fmt.Sprintf("%s and %s resulted in a draw\n", winner, loser)
			err := publishGameLog(gl, amqpChan)
			if err != nil {
				fmt.Printf("error publishing war message: %s", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("Error: wrong HandleWar outcome")
			return pubsub.NackDiscard
		}
	}
}
