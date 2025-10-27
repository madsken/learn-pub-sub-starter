package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	amqpChan, amqpQueue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return err
	}

	amqpDelivery, err := amqpChan.Consume(
		amqpQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		defer amqpChan.Close()
		for message := range amqpDelivery {
			var result T
			err = json.Unmarshal(message.Body, &result)
			if err != nil {
				fmt.Printf("Could not unmarshal message: %v\n", err)
				continue
			}
			ackType := handler(result)
			switch ackType {
			case Ack:
				message.Ack(false)
			case NackRequeue:
				message.Nack(false, true)
			case NackDiscard:
				message.Nack(false, false)
			}
		}
	}()

	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	amqpChan, amqpQueue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return err
	}

	amqpDelivery, err := amqpChan.Consume(
		amqpQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		defer amqpChan.Close()
		for message := range amqpDelivery {
			var result T
			buf := bytes.NewBuffer(message.Body)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&result)
			if err != nil {
				fmt.Println("Error decoding gob")
				continue
			}

			ackType := handler(result)
			switch ackType {
			case Ack:
				message.Ack(false)
			case NackRequeue:
				message.Nack(false, true)
			case NackDiscard:
				message.Nack(false, false)
			}
		}
	}()

	return nil
}
