package repository

import (
	"context"
	"errors"
	"listenly-backend/internal/room/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")
var ErrAlreadyMember = errors.New("user already a member of this room")

type RoomRepository interface {
	CreateRoom(ctx context.Context, room *models.Room) error
	FindRoomByUUID(ctx context.Context, u uuid.UUID) (*models.Room, error)
	FindRoomByName(ctx context.Context, name string) (*models.Room, error)
	FindRoomByInviteCode(ctx context.Context, code string) (*models.Room, error)
	FindRoomByInviteToken(ctx context.Context, token string) (*models.Room, error)
	CountMembers(ctx context.Context, roomID int64) (int64, error)
	AddMember(ctx context.Context, member *models.RoomMember) error
	IsMember(ctx context.Context, roomID, userID int64) (bool, error)
	RemoveMember(ctx context.Context, roomID, userID int64) error
	ListPublicRooms(ctx context.Context, limit, offset int) ([]models.Room, int64, error)
	ListRoomsByMemberUserID(ctx context.Context, userID int64, limit, offset int) ([]models.Room, int64, error)
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) CreateRoom(ctx context.Context, room *models.Room) error {
	if room.UUID == uuid.Nil {
		v7, err := uuid.NewV7()
		if err != nil {
			return err
		}
		room.UUID = v7
	}
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *roomRepository) FindRoomByUUID(ctx context.Context, u uuid.UUID) (*models.Room, error) {
	var room models.Room
	err := r.db.WithContext(ctx).First(&room, "uuid = ?", u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &room, err
}

func (r *roomRepository) FindRoomByName(ctx context.Context, name string) (*models.Room, error) {
	var room models.Room
	err := r.db.WithContext(ctx).First(&room, "LOWER(name) = ?", strings.ToLower(name)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &room, err
}

func (r *roomRepository) FindRoomByInviteCode(ctx context.Context, code string) (*models.Room, error) {
	var room models.Room
	err := r.db.WithContext(ctx).First(&room, "invite_code = ?", code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &room, err
}

func (r *roomRepository) FindRoomByInviteToken(ctx context.Context, token string) (*models.Room, error) {
	var room models.Room
	err := r.db.WithContext(ctx).First(&room, "invite_token = ?", token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &room, err
}

func (r *roomRepository) CountMembers(ctx context.Context, roomID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RoomMember{}).Where("room_id = ?", roomID).Count(&count).Error
	return count, err
}

func (r *roomRepository) AddMember(ctx context.Context, member *models.RoomMember) error {
	isMember, err := r.IsMember(ctx, member.RoomID, member.UserID)
	if err != nil {
		return err
	}
	if isMember {
		return ErrAlreadyMember
	}
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *roomRepository) IsMember(ctx context.Context, roomID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).Count(&count).Error
	return count > 0, err
}

func (r *roomRepository) RemoveMember(ctx context.Context, roomID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&models.RoomMember{}).Error
}

func (r *roomRepository) ListPublicRooms(ctx context.Context, limit, offset int) ([]models.Room, int64, error) {
	var rooms []models.Room
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.Room{}).
		Where("visibility = ?", models.VisibilityPublic).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("visibility = ?", models.VisibilityPublic).
		Limit(limit).Offset(offset).
		Find(&rooms).Error

	return rooms, total, err
}

func (r *roomRepository) ListRoomsByMemberUserID(ctx context.Context, userID int64, limit, offset int) ([]models.Room, int64, error) {
	var rooms []models.Room
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Room{}).
		Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Where("room_members.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Find(&rooms).Error
	return rooms, total, err
}
