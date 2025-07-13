package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
    defer fmt.Println("Peril server shutting down")

    rabbitStr := "amqp://guest:guest@localhost:5672/"
    conn, err := amqp.Dial(rabbitStr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    defer fmt.Println("\nDisconnected from broker")
    fmt.Println("Connection successful")

    ch, err := conn.Channel()
    if err != nil {
        log.Fatal(err)
    }


    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT)
    <-sigs

    gamelogic.PrintServerHelp()
    for {
        cmds := gamelogic.GetInput()
        if len(cmds) = 0 {
            continue
        }
        cmd = cmds[0]

        switch cmd {
        case "pause":
            err = pubsub.PublishJSON(
                ch, 
                string(routing.ExchangePerilDirect), 
                string(routing.PauseKey), 
                routing.PlayingState{IsPaused: true},
            )
            if err != nil {
                log.Fatal(err)
            }
        case "resume":
            err = pubsub.PublishJSON(
                ch, 
                string(routing.ExchangePerilDirect), 
                string(routing.PauseKey), 
                routing.PlayingState{IsPaused: false},
            )
            if err != nil {
                log.Fatal(err)
            }
        case "quit":
            fmt.Println("exiting")
            break
        default:
            fmt.Println("command not understood")
        }
    }
}
