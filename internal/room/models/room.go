package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	UUID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Name        string    `gorm:"not null"`
	HostID      int64     `gorm:"not null;index"`
	Visibility  string    `gorm:"not null;default:public"`
	InviteCode  *string   `gorm:"uniqueIndex"`
	InviteToken *string   `gorm:"uniqueIndex"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Room) TableName() string {
	return "rooms"
}

const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)
