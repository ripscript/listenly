package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;type:bigint"`
	UUID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.UUID == uuid.Nil {
		uuidV7, err := uuid.NewV7()
		if err != nil {
			return err
		}
		base.UUID = uuidV7
	}
	return
}
