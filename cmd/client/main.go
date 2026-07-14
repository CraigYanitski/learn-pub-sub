package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    fmt.Println("Starting Peril client...")
    defer fmt.Println("Peril client shutting down")

	// establish connection with rabbitmq server
    rabbitStr := "amqp://guest:guest@localhost:5672/"
    conn, err := amqp.Dial(rabbitStr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    defer fmt.Println("\nDisconnected from broker")
    fmt.Println("Connection successful")

	// Create channel on connection
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	// Get client username
    username, err := gamelogic.ClientWelcome()
    if err != nil {
        log.Fatal(err)
    }

	// Create game state for client
    gamestate := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(
		conn, 
		string(routing.ExchangePerilDirect), 
		strings.Join([]string{routing.PauseKey, username}, "."), 
		routing.PauseKey, 
		pubsub.Transient, 
		handlerPause(gamestate),
	)
    if err != nil {
        log.Fatal(err)
    }
	err = pubsub.SubscribeJSON(
		conn, 
		routing.ExchangePerilTopic, 
		strings.Join([]string{routing.ArmyMovesPrefix, username}, "."), 
		strings.Join([]string{routing.ArmyMovesPrefix,      "*"}, "."), 
		pubsub.Transient, 
		handlerMove(gamestate, ch),
	)
    if err != nil {
        log.Fatal(err)
    }
	err = pubsub.SubscribeJSON(
		conn, 
		routing.ExchangePerilTopic, 
		routing.WarRecognitionsPrefix, 
		routing.WarRecognitionsPrefix+".*",
		//fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix), 
		pubsub.Durable, 
		handlerWar(gamestate),
	)
    if err != nil {
        log.Fatal(err)
    }

	// Listen for SIGINT
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT)

	// create app command input channel
    cmdsChan := make(chan []string)
    go func() {
        for {
            cmds := gamelogic.GetInput()
            cmdsChan <- cmds
			time.Sleep(10 * time.Millisecond)
        }
    }()

    cli:
    for {
		// race between SIGINT and app command
        select {
        case <-sigs:
            fmt.Println("Interrupt: Program is shutting down")
            break cli
        case cmds, ok := <-cmdsChan:
			// Exit app if channel closed
            if !ok {
                fmt.Println("Interrupt: Program is shutting down")
                break cli
            }
            if len(cmds) == 0 {
                continue
            }
            cmd := cmds[0]

            switch cmd {
            case "spawn":
                err = gamestate.CommandSpawn(cmds)
                if err != nil {
                    fmt.Println("Not a valid spawn command: spawn [location] [unit]")
                }
            case "move":
                move, err := gamestate.CommandMove(cmds)
                if err != nil {
                    fmt.Println("Not a valid move command: move [location] [ID]")
                }
				pubsub.PublishJSON(
					ch, 
					string(routing.ExchangePerilTopic), 
					strings.Join([]string{routing.ArmyMovesPrefix, username}, "."), 
					move,
				)
				fmt.Println("Published move successfully")
            case "status":
                gamestate.CommandStatus()
            case "help":
                gamelogic.PrintClientHelp()
            case "spam":
                fmt.Println("Spamming not allowed yet!")
            case "quit":
                fmt.Println("exiting game")
                break cli
            default:
                fmt.Println("command not understood")
            }
        }
    }
}
