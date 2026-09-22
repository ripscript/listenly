package models

import "time"

type RoomMember struct {
	ID       int64     `gorm:"primaryKey;autoIncrement"`
	RoomID   int64     `gorm:"not null;index"`
	UserID   int64     `gorm:"not null;index"`
	Role     string    `gorm:"not null;default:member"`
	JoinedAt time.Time `gorm:"autoCreateTime"`
}

func (RoomMember) TableName() string {
	return "room_members"
}

const (
	RoleHost   = "host"
	RoleMember = "member"
)
