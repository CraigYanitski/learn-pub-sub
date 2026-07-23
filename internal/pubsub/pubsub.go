package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
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

	table := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",	
	}
    qu, err := ch.QueueDeclare(queue, durable, autodelete, exclusive, false, table)
    if err != nil {
        return nil, amqp.Queue{}, err
    }

	err = ch.QueueBind(queue, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

    return ch, qu, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
)  error {
	ch, qu, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveryChan, err := ch.Consume(qu.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func(){
		defer ch.Close()
		for d := range deliveryChan {
			var delivery T
			err := json.Unmarshal(d.Body, &delivery)
			if err != nil {
				//d.Ack(false)
				print(err)
				continue
			}
			acktype := handler(delivery)
			switch acktype {
			case Ack:
				//log.Println("message delivery acknowledged")
				d.Ack(false)
			case NackRequeue:
				//log.Println("message delivery failed, re-queuing")
				d.Nack(false, true)
			case NackDiscard:
				//log.Println("message delivery failed, discarding")
				d.Nack(false, false)
			}
		}
	}()

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}
	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body: buffer.Bytes(),
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if err != nil {
		return err
	}
	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection, 
	exchange, 
	queueName, 
	key string, 
	queueType SimpleQueueType, 
	handler func(T) AckType, 
	unmarshaller func([]byte) (T, error),
) error {
	ch, qu, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveryChan, err := ch.Consume(qu.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func(){
		defer ch.Close()
		for d := range deliveryChan {
			delivery, err := unmarshaller(d.Body)
			if err != nil {
				log.Println(err)
				continue
			}
			acktype := handler(delivery)
			switch acktype {
			case Ack:
				//log.Println("message delivery acknowledged")
				d.Ack(false)
			case NackRequeue:
				//log.Println("message delivery failed, re-queuing")
				d.Nack(false, true)
			case NackDiscard:
				//log.Println("message delivery failed, discarding")
				d.Nack(false, false)
			}
		}
	}()

	return nil
}

