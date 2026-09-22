package models

import (
	"time"

	"github.com/google/uuid"
)

type Track struct {
	ID              int64     `gorm:"primaryKey;autoIncrement"`
	UUID            uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	YoutubeVideoID  string    `gorm:"column:youtube_video_id;not null;uniqueIndex"`
	Title           string    `gorm:"not null"`
	Artist          string
	DurationSeconds int    `gorm:"column:duration_seconds;not null;default:0"`
	ThumbnailURL    string `gorm:"column:thumbnail_url"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Track) TableName() string {
	return "tracks"
}
