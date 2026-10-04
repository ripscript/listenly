package rabbitmq

import (
	"fmt"
	"time"
)

// Tipe-tipe event yang mengalir di listenly.events.
const (
	EventPlaybackUpdated = "playback_updated"
	EventQueueUpdated    = "queue_updated"
	EventPresenceUpdated = "presence_updated"
)

// Event adalah amplop seragam untuk semua pesan yang di-publish.
type Event struct {
	Type      string    `json:"type"`
	RoomUUID  string    `json:"room_uuid"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// Routing key helpers — pola: room.<uuid>.<kategori>
func RoomPlaybackKey(roomUUID string) string { return fmt.Sprintf("room.%s.playback", roomUUID) }
func RoomQueueKey(roomUUID string) string    { return fmt.Sprintf("room.%s.queue", roomUUID) }
func RoomPresenceKey(roomUUID string) string { return fmt.Sprintf("room.%s.presence", roomUUID) }
