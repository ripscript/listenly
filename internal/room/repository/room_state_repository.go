package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RoomState struct {
	CurrentTrackID  string    `json:"current_track_id"`
	PositionSeconds int       `json:"position_seconds"`
	IsPlaying       bool      `json:"is_playing"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RoomStateRepository interface {
	InitState(ctx context.Context, roomUUID string) error
	GetState(ctx context.Context, roomUUID string) (*RoomState, error)
	UpdateState(ctx context.Context, roomUUID string, state RoomState) error
	DeleteState(ctx context.Context, roomUUID string) error

	AddOnlineMember(ctx context.Context, roomUUID, userUUID string) error
	RemoveOnlineMember(ctx context.Context, roomUUID, userUUID string) error
	CountOnlineMembers(ctx context.Context, roomUUID string) (int64, error)
	IsOnline(ctx context.Context, roomUUID, userUUID string) (bool, error)
}

type roomStateRepository struct {
	rdb *redis.Client
}

func NewRoomStateRepository(rdb *redis.Client) RoomStateRepository {
	return &roomStateRepository{rdb: rdb}
}

func stateKey(roomUUID string) string {
	return "room:" + roomUUID + ":state"
}

func onlineKey(roomUUID string) string {
	return "room:" + roomUUID + ":online"
}

func (r *roomStateRepository) InitState(ctx context.Context, roomUUID string) error {
	state := RoomState{
		IsPlaying: false,
		UpdatedAt: time.Now(),
	}
	return r.UpdateState(ctx, roomUUID, state)
}

func (r *roomStateRepository) GetState(ctx context.Context, roomUUID string) (*RoomState, error) {
	val, err := r.rdb.Get(ctx, stateKey(roomUUID)).Result()
	if err == redis.Nil {
		return nil, nil // state belum ada, bukan error
	}
	if err != nil {
		return nil, err
	}

	var state RoomState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *roomStateRepository) UpdateState(ctx context.Context, roomUUID string, state RoomState) error {
	state.UpdatedAt = time.Now()
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, stateKey(roomUUID), data, 0).Err()
}

func (r *roomStateRepository) DeleteState(ctx context.Context, roomUUID string) error {
	return r.rdb.Del(ctx, stateKey(roomUUID), onlineKey(roomUUID)).Err()
}

func (r *roomStateRepository) AddOnlineMember(ctx context.Context, roomUUID, userUUID string) error {
	return r.rdb.SAdd(ctx, onlineKey(roomUUID), userUUID).Err()
}

func (r *roomStateRepository) RemoveOnlineMember(ctx context.Context, roomUUID, userUUID string) error {
	return r.rdb.SRem(ctx, onlineKey(roomUUID), userUUID).Err()
}

func (r *roomStateRepository) CountOnlineMembers(ctx context.Context, roomUUID string) (int64, error) {
	return r.rdb.SCard(ctx, onlineKey(roomUUID)).Result()
}

func (r *roomStateRepository) IsOnline(ctx context.Context, roomUUID, userUUID string) (bool, error) {
	return r.rdb.SIsMember(ctx, onlineKey(roomUUID), userUUID).Result()
}
