package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
    Transient = iota
    Durable
)

func main() {
	fmt.Println("Starting Peril client...")
    defer fmt.Println("Peril client shutting down")

    rabbitStr := "amqp://guest:guest@localhost:5672/"
    conn, err := amqp.Dial(rabbitStr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    defer fmt.Println("\nDisconnected from broker")
    fmt.Println("Connection successful")

    username, err := gamelogic.ClientWelcome()
    if err != nil {
        log.Fatal(err)
    }

    _, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, strings.Join([]string{routing.PauseKey, username}, "."), routing.PauseKey, Transient)
    if err != nil {
        log.Fatal(err)
    }

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT)
    <-sigs
}
