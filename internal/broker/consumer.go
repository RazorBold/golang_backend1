package broker

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(ctx context.Context, body []byte) error

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// max 10 unacked messages per consumer
	if err := ch.Qos(10, 0, false); err != nil {
		conn.Close()
		return nil, err
	}

	if err := SetupTopology(ch); err != nil {
		conn.Close()
		return nil, err
	}

	return &Consumer{conn: conn, ch: ch}, nil
}

// Consume blocks dan memproses pesan sampai ctx dibatalkan.
// Jika handler error → nack tanpa requeue (masuk DLQ).
func (c *Consumer) Consume(ctx context.Context, queue string, handler MessageHandler) error {
	msgs, err := c.ch.ConsumeWithContext(ctx, queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("rabbitmq channel closed")
			}
			if err := handler(ctx, msg.Body); err != nil {
				msg.Nack(false, false) // ke DLQ
				continue
			}
			msg.Ack(false)
		}
	}
}

func (c *Consumer) Close() {
	c.ch.Close()
	c.conn.Close()
}
