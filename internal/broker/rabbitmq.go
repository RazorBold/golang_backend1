package broker

import amqp "github.com/rabbitmq/amqp091-go"

const (
	ExchangeTelemetry = "iot.telemetry"
	ExchangeDLX       = "iot.dlx"
	QueuePersist      = "telemetry.persist"
	QueueDLQ          = "dlq.telemetry"
	RoutingKeyAll     = "telemetry.#"
)

// SetupTopology mendeklarasikan semua exchange, queue, dan binding.
// Dipanggil baik oleh publisher maupun worker agar idempotent.
func SetupTopology(ch *amqp.Channel) error {
	// Dead Letter Exchange
	if err := ch.ExchangeDeclare(ExchangeDLX, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(QueueDLQ, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(QueueDLQ, QueueDLQ, ExchangeDLX, false, nil); err != nil {
		return err
	}

	// Main telemetry exchange (topic)
	if err := ch.ExchangeDeclare(ExchangeTelemetry, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	// Main queue dengan DLX fallback
	args := amqp.Table{
		"x-dead-letter-exchange":    ExchangeDLX,
		"x-dead-letter-routing-key": QueueDLQ,
	}
	if _, err := ch.QueueDeclare(QueuePersist, true, false, false, false, args); err != nil {
		return err
	}

	return ch.QueueBind(QueuePersist, RoutingKeyAll, ExchangeTelemetry, false, nil)
}
