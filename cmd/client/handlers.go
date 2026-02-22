package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {

	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerArmyMoves(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(m gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(m)
		if outcome == gamelogic.MoveOutComeSafe {
			return pubsub.Ack
		}
		if outcome == gamelogic.MoveOutcomeMakeWar {
			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername()),
				gamelogic.RecognitionOfWar{
					Attacker: m.Player,
					Defender: gs.GetPlayerSnap(),
				})
			if err != nil {
				return pubsub.NackReque
			}
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	}
}

func handlerWarMoves(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(msg gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(msg)
		if outcome == gamelogic.WarOutcomeNotInvolved {
			return pubsub.NackReque
		}
		if outcome == gamelogic.WarOutcomeNoUnits {
			return pubsub.NackDiscard
		}
		if outcome == gamelogic.WarOutcomeYouWon || outcome == gamelogic.WarOutcomeOpponentWon {
			log := fmt.Sprintf("%s won a war against %s", winner, loser)
			return publishGamelogs(gs, ch, log)
		}
		if outcome == gamelogic.WarOutcomeDraw {
			log := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			return publishGamelogs(gs, ch, log)
		}
		fmt.Println("Error processing war outcome")
		return pubsub.NackDiscard
	}
}

func publishGamelogs(gs *gamelogic.GameState, ch *amqp.Channel, logMessage string) pubsub.AckType {
	gamelog := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     logMessage,
		Username:    gs.GetUsername(),
	}

	err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername()), gamelog)
	if err != nil {
		fmt.Println("Error publishing gamelog:", err)
		return pubsub.NackReque
	}
	return pubsub.Ack
}
