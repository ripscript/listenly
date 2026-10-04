package repository

import (
	"context"
	"errors"
	"listenly-backend/internal/music/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

type MusicRepository interface {
	SearchTracks(ctx context.Context, query string, limit, offset int) ([]models.Track, int64, error)
	FindTrackByUUID(ctx context.Context, u uuid.UUID) (*models.Track, error)
	FindTrackByYoutubeID(ctx context.Context, videoID string) (*models.Track, error)
	CreateTrack(ctx context.Context, track *models.Track) error

	CreateQueueItem(ctx context.Context, item *models.QueueItem) error
	FindQueueItemByUUID(ctx context.Context, u uuid.UUID) (*models.QueueItem, error)
	ListQueueByRoomID(ctx context.Context, roomID int64) ([]models.QueueItem, error)
	CountQueueByRoomID(ctx context.Context, roomID int64) (int64, error)
	DeleteQueueItem(ctx context.Context, id int64) error
	UpdateQueueItemStatus(ctx context.Context, id int64, status string) error
	FindTracksByUUIDs(ctx context.Context, uuids []string) ([]models.Track, error)
	FindNextReadyInRoom(ctx context.Context, roomID int64) (*models.QueueItem, error)
}

type musicRepository struct {
	db *gorm.DB
}

func NewMusicRepository(db *gorm.DB) MusicRepository {
	return &musicRepository{db: db}
}

func (r *musicRepository) SearchTracks(ctx context.Context, query string, limit, offset int) ([]models.Track, int64, error) {
	var tracks []models.Track
	var total int64

	likeQuery := "%" + strings.ToLower(query) + "%"

	base := r.db.WithContext(ctx).Model(&models.Track{}).
		Where("LOWER(title) LIKE ? OR LOWER(artist) LIKE ?", likeQuery, likeQuery)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := base.Limit(limit).Offset(offset).Find(&tracks).Error
	return tracks, total, err
}

func (r *musicRepository) FindTrackByUUID(ctx context.Context, u uuid.UUID) (*models.Track, error) {
	var track models.Track
	err := r.db.WithContext(ctx).First(&track, "uuid = ?", u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &track, err
}

func (r *musicRepository) FindTrackByYoutubeID(ctx context.Context, videoID string) (*models.Track, error) {
	var track models.Track
	err := r.db.WithContext(ctx).First(&track, "youtube_video_id = ?", videoID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &track, err
}

func (r *musicRepository) CreateTrack(ctx context.Context, track *models.Track) error {
	if track.UUID == uuid.Nil {
		v7, err := uuid.NewV7()
		if err != nil {
			return err
		}
		track.UUID = v7
	}
	return r.db.WithContext(ctx).Create(track).Error
}

func (r *musicRepository) CreateQueueItem(ctx context.Context, item *models.QueueItem) error {
	if item.UUID == uuid.Nil {
		v7, err := uuid.NewV7()
		if err != nil {
			return err
		}
		item.UUID = v7
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *musicRepository) FindQueueItemByUUID(ctx context.Context, u uuid.UUID) (*models.QueueItem, error) {
	var item models.QueueItem
	err := r.db.WithContext(ctx).Preload("Track").First(&item, "uuid = ?", u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &item, err
}

func (r *musicRepository) ListQueueByRoomID(ctx context.Context, roomID int64) ([]models.QueueItem, error) {
	var items []models.QueueItem
	err := r.db.WithContext(ctx).Preload("Track").
		Where("room_id = ? AND status != ?", roomID, models.StatusPlayed).
		Order("position ASC").
		Find(&items).Error
	return items, err
}

func (r *musicRepository) CountQueueByRoomID(ctx context.Context, roomID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.QueueItem{}).
		Where("room_id = ? AND status != ?", roomID, models.StatusPlayed).
		Count(&count).Error
	return count, err
}

func (r *musicRepository) DeleteQueueItem(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.QueueItem{}, id).Error
}

func (r *musicRepository) UpdateQueueItemStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).Model(&models.QueueItem{}).
		Where("id = ?", id).Update("status", status).Error
}

func (r *musicRepository) FindTracksByUUIDs(ctx context.Context, uuids []string) ([]models.Track, error) {
	var tracks []models.Track
	err := r.db.WithContext(ctx).Where("uuid IN ?", uuids).Find(&tracks).Error
	return tracks, err
}

func (r *musicRepository) FindNextReadyInRoom(ctx context.Context, roomID int64) (*models.QueueItem, error) {
	var item models.QueueItem
	err := r.db.WithContext(ctx).Preload("Track").
		Where("room_id = ? AND status = ?", roomID, models.StatusReady).
		Order("position ASC").
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &item, err
}
