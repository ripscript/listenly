package ws

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/coder/websocket/wsjson"

	"listenly-backend/pkg/rabbitmq"
)

// StartDispatcher menjalankan loop yang meng-consume event dari RabbitMQ
// dan mem-push-nya ke semua koneksi WebSocket di room terkait.
// Dijalankan sebagai goroutine di main.go.
func StartDispatcher(hub *Hub, consumer *rabbitmq.Consumer) error {
	deliveries, err := consumer.Consume()
	if err != nil {
		return err
	}

	go func() {
		for d := range deliveries {
			var event rabbitmq.Event
			if err := json.Unmarshal(d.Body, &event); err != nil {
				slog.Warn("failed to decode event", "error", err)
				continue
			}
			if event.RoomUUID == "" {
				continue
			}

			conns := hub.Connections(event.RoomUUID)
			for _, conn := range conns {
				// best-effort push; abaikan error per-koneksi (client mungkin baru putus)
				_ = wsjson.Write(context.Background(), conn, event)
			}
		}
		slog.Warn("rabbitmq dispatcher stopped (deliveries channel closed)")
	}()

	return nil
}
