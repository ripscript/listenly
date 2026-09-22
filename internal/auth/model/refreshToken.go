package models

import (
	"time"
)

type RefreshToken struct {
	BaseModel
	UserID    uint      `gorm:"type:bigint;index:idx_refresh_tokens_user_id;not null"`
	Token     string    `gorm:"type:varchar(255);index:idx_refresh_tokens_token;not null"`
	Revoked   bool      `gorm:"not null;default:false"`
	ExpiresAt time.Time `gorm:"type:timestamptz;not null"`

	User User `gorm:"foreignKey:UserID"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
