package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT)
    <-sigs
}
