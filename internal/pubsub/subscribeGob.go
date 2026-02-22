package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
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
			buf := bytes.NewBuffer(d.Body)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&val)
			if err != nil {
				continue
			}
			fmt.Printf("received message on queue %s\n", queue.Name)
			acknowledged := handler(val)
			switch acknowledged {
			case Ack:
				d.Ack(false)
				fmt.Println("message acked")
			case NackReque:
				d.Nack(false, true)
				fmt.Println("message nacked requeue")
			case NackDiscard:
				d.Nack(false, false)
				fmt.Println("message nacked discard")
			}
		}

	}()

	return nil
}
