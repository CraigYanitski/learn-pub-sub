package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
    body, err := json.Marshal(val)
    if err != nil {
        return err
    }
    msg := amqp.Publishing{
        ContentType: "application/json",
        Body: body,
    }
    err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
    if err != nil {
        return err
    }

    return nil
}

type SimpleQueueType int

const (
    Transient = iota
    Durable
)

func DeclareAndBind(conn *amqp.Connection, exchange, queue, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
    ch, err := conn.Channel()
    if err != nil {
        return nil, amqp.Queue{}, err
    }

    var (
        durable bool
        autodelete bool
        exclusive bool
    )

    switch queueType {
    case Durable:
        durable = true
    case Transient:
        autodelete = true
        exclusive = true
    }

    qu, err := ch.QueueDeclare(queue, durable, autodelete, exclusive, false, nil)
    if err != nil {
        return nil, amqp.Queue{}, err
    }


    return ch, qu, nil
}

