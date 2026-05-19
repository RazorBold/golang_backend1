package broker

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// publisher confirms: API tidak return 201 sebelum broker ack
	if err := ch.Confirm(false); err != nil {
		conn.Close()
		return nil, err
	}

	if err := SetupTopology(ch); err != nil {
		conn.Close()
		return nil, err
	}

	return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) PublishTelemetry(ctx context.Context, msg TelemetryMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := fmt.Sprintf("telemetry.%s.%s", msg.AppID, msg.DeviceID)

	dc, err := p.ch.PublishWithDeferredConfirmWithContext(ctx,
		ExchangeTelemetry, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	confirmed, err := dc.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !confirmed {
		return fmt.Errorf("broker nacked the message")
	}
	return nil
}

func (p *Publisher) Close() {
	p.ch.Close()
	p.conn.Close()
}
