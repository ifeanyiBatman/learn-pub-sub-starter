package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {


	var gobBytes bytes.Buffer
	enc := gob.NewEncoder(&gobBytes)
	err := enc.Encode(val)
	if err != nil {
			return err
	}

	err = ch.PublishWithContext(context.Background(),exchange,key,false,false,amqp.Publishing{
		ContentType: "application/gob",
		Body: gobBytes.Bytes(),
	})
	if err != nil {
		return err
	}
	return nil
}