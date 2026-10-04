package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer mendengarkan event dari exchange listenly.events.
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewConsumer membuka koneksi, mendeklarasikan exchange (idempoten),
// membuat queue sementara eksklusif, dan bind ke pola routing key tertentu.
func NewConsumer(url string, bindingKeys []string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// queue eksklusif & auto-delete: tiap instance gateway punya queue sendiri,
	// hilang saat gateway mati. Cocok untuk fan-out realtime (bukan work-queue).
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	for _, key := range bindingKeys {
		if err := ch.QueueBind(q.Name, key, ExchangeName, false, nil); err != nil {
			ch.Close()
			conn.Close()
			return nil, err
		}
	}

	return &Consumer{conn: conn, channel: ch, queue: q.Name}, nil
}

// Consume mengembalikan channel pesan yang bisa di-range.
func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	return c.channel.Consume(c.queue, "", true, false, false, false, nil)
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
