package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Transient SimpleQueueType = iota
	Durable
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	amqpChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable, autoDelete, exclusive := getParametersFromQueueType(queueType)
	amqpQueue, err := amqpChan.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, nil
	}

	err = amqpChan.QueueBind(amqpQueue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, nil
	}

	return amqpChan, amqpQueue, nil
}

func getParametersFromQueueType(queueType SimpleQueueType) (bool, bool, bool) {
	if queueType == Transient {
		return false, true, true
	}
	return true, false, false
}
