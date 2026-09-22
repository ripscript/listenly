package models

import (
	"time"

	"github.com/google/uuid"
)

type QueueItem struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	UUID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	RoomID      int64     `gorm:"not null;index"`
	RoomHostID  int64     `gorm:"not null"`
	TrackID     int64     `gorm:"not null"`
	Track       Track     `gorm:"foreignKey:TrackID"`
	RequestedBy int64     `gorm:"not null"`
	Status      string    `gorm:"not null;default:pending"`
	Position    int       `gorm:"not null;default:0"`
	CreatedAt   time.Time
}

func (QueueItem) TableName() string {
	return "queue_items"
}

const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusFailed  = "failed"
	StatusPlayed  = "played"
)
