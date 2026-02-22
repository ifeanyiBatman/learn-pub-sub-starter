package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	channel, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()

	gamestate := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, username), routing.PauseKey, pubsub.Transient, handlerPause(gamestate))
	if err != nil {
		fmt.Printf("could not subscribe to pause: %v\n", err)
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix), pubsub.Transient, handlerArmyMoves(gamestate, channel))
	if err != nil {
		fmt.Printf("could not subscribe to army moves: %v\n", err)
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix), pubsub.Durable, handlerWarMoves(gamestate, channel))
	if err != nil {
		fmt.Printf("could not subscribe to war: %v\n", err)
	}
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		if words[0] == "spawn" {
			if err := gamestate.CommandSpawn(words); err != nil {
				fmt.Println(err)
			}
			continue
		}
		if words[0] == "move" {
			move, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(channel, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username), move)
			if err != nil {
				fmt.Printf("error publishing move: %v\n", err)
				continue
			}
			fmt.Println("Move succesfully published")
			continue
		}
		if words[0] == "status" {
			gamestate.CommandStatus()
			continue
		}
		if words[0] == "help" {
			gamelogic.PrintClientHelp()
			continue
		}
		if words[0] == "spam" {
			fmt.Println("Spamming not allowed yet!")
			continue
		}
		if words[0] == "quit" {
			gamelogic.PrintQuit()
			return
		}
		fmt.Println("invalid command")
		continue

	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nShutting down Peril server...")
}
