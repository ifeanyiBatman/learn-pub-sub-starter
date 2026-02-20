package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	delivery, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range delivery {
			var val T
			err := json.Unmarshal(d.Body, &val)
			if err != nil {
				continue
			}
			fmt.Printf("received message on queue %s\n", queue.Name)
			handler(val)
			d.Ack(false)
		}

	}()

	return nil
}
