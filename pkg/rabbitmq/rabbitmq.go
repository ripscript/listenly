package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ExchangeName adalah topic exchange utama untuk semua event Listenly.
const ExchangeName = "listenly.events"

// Publisher membungkus koneksi + channel untuk mem-publish event.
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewPublisher membuka koneksi ke RabbitMQ, mendeklarasikan topic exchange,
// dan mengembalikan publisher siap pakai.
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

	// topic exchange: routing fleksibel berbasis pola (mis. "room.<uuid>.playback")
	if err := ch.ExchangeDeclare(
		ExchangeName,
		"topic",
		true,  // durable: bertahan setelah broker restart
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Publisher{conn: conn, channel: ch}, nil
}

// Publish menyerialisasi payload ke JSON dan mem-publish-nya dengan routing key.
func (p *Publisher) Publish(ctx context.Context, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(
		ctx,
		ExchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // pesan bertahan di queue durable
			Timestamp:    time.Now(),
		},
	)
}

// Close menutup channel dan koneksi dengan rapi.
func (p *Publisher) Close() error {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
