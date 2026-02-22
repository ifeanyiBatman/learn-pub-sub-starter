package pubsub

import amqp "github.com/rabbitmq/amqp091-go"

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // SimpleQueueType is an "enum" type I made to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	isDurable := queueType == "durable"
	isAutoDelete := !isDurable
	isExclusive := isAutoDelete

	queue, err := channel.QueueDeclare(queueName, isDurable, isAutoDelete, isExclusive, false, amqp.Table{"x-dead-letter-exchange": "peril_dlx"})
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return channel, queue, nil

}
