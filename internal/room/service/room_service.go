package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"time"

	authv1 "listenly-backend/gen/go/auth/v1"
	"listenly-backend/internal/room/dto"
	"listenly-backend/internal/room/models"
	"listenly-backend/internal/room/repository"
)

var (
	ErrRoomNotFound  = errors.New("room not found")
	ErrInvalidCode   = errors.New("invalid invite code")
	ErrInvalidToken  = errors.New("invalid invite token")
	ErrNotAuthorized = errors.New("not authorized")
	ErrUserNotFound  = errors.New("user not found")
)

type RoomService interface {
	CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (*dto.RoomResponse, error)
	GetRoom(ctx context.Context, roomUUID, requesterUUID string) (*dto.RoomResponse, error)
	JoinRoomByCode(ctx context.Context, req dto.JoinRoomByCodeRequest) (*dto.RoomResponse, error)
	JoinRoomByToken(ctx context.Context, req dto.JoinRoomByTokenRequest) (*dto.RoomResponse, error)
	LeaveRoom(ctx context.Context, roomUUID, userUUID string) error
	ListPublicRooms(ctx context.Context, page, pageSize int) ([]dto.RoomResponse, int64, error)
	JoinRoom(ctx context.Context, req dto.JoinRoomRequest) (*dto.RoomResponse, error)
	ListMyRooms(ctx context.Context, userUUID string, page, pageSize int) ([]dto.RoomResponse, int64, error)
	GetRoomInternal(ctx context.Context, roomUUID string) (int64, int64, error)
	CheckMembership(ctx context.Context, roomUUID, userUUID string) (isMember bool, isHost bool, err error)
	UpdatePlayback(ctx context.Context, roomUUID, requesterUUID, currentTrackID string, positionSeconds int, isPlaying bool) (*PlaybackState, error)
	GetPlayback(ctx context.Context, roomUUID, requesterUUID string) (*PlaybackState, error)
}

type roomService struct {
	repo       repository.RoomRepository
	stateRepo  repository.RoomStateRepository
	authClient authv1.AuthServiceClient
}

func NewRoomService(repo repository.RoomRepository, stateRepo repository.RoomStateRepository, authClient authv1.AuthServiceClient) RoomService {
	return &roomService{repo: repo, stateRepo: stateRepo, authClient: authClient}
}

func (s *roomService) resolveUserID(ctx context.Context, userUUID string) (int64, error) {
	resp, err := s.authClient.GetUserByUUID(ctx, &authv1.GetUserByUUIDRequest{Uuid: userUUID})
	if err != nil {
		return 0, ErrUserNotFound
	}
	return resp.Id, nil
}

func (s *roomService) CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (*dto.RoomResponse, error) {
	hostID, err := s.resolveUserID(ctx, req.HostUUID)
	if err != nil {
		return nil, err
	}

	room := &models.Room{
		Name:       req.Name,
		HostID:     hostID,
		Visibility: req.Visibility,
	}

	if req.Visibility == models.VisibilityPrivate {
		code, err := generateInviteCode()
		if err != nil {
			return nil, err
		}
		token, err := generateInviteToken()
		if err != nil {
			return nil, err
		}
		room.InviteCode = &code
		room.InviteToken = &token
	}

	getRoomByName, err := s.repo.FindRoomByName(ctx, req.Name)
	if err == nil && getRoomByName != nil {
		return nil, errors.New("room name already exists")
	}

	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}

	if err := s.repo.AddMember(ctx, &models.RoomMember{
		RoomID: room.ID,
		UserID: hostID,
		Role:   models.RoleHost,
	}); err != nil {
		return nil, err
	}

	if err := s.stateRepo.InitState(ctx, room.UUID.String()); err != nil {
		return nil, err
	}

	return s.toResponse(ctx, room, req.HostUUID, true)
}

func (s *roomService) GetRoom(ctx context.Context, roomUUID, requesterUUID string) (*dto.RoomResponse, error) {
	room, err := s.repo.FindRoomByUUID(ctx, mustParseUUID(roomUUID))
	if err != nil {
		return nil, ErrRoomNotFound
	}

	if room.Visibility == models.VisibilityPrivate {
		requesterID, err := s.resolveUserID(ctx, requesterUUID)
		if err != nil {
			return nil, ErrRoomNotFound
		}

		isMember, err := s.repo.IsMember(ctx, room.ID, requesterID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrRoomNotFound
		}
	}

	isHost, err := s.isRequesterHost(ctx, room, requesterUUID)
	if err != nil {
		return nil, err
	}

	return s.toResponse(ctx, room, requesterUUID, isHost)
}

func (s *roomService) JoinRoomByCode(ctx context.Context, req dto.JoinRoomByCodeRequest) (*dto.RoomResponse, error) {
	room, err := s.repo.FindRoomByInviteCode(ctx, req.InviteCode)
	if err != nil {
		return nil, ErrInvalidCode
	}
	return s.joinRoom(ctx, room, req.UserUUID)
}

func (s *roomService) JoinRoomByToken(ctx context.Context, req dto.JoinRoomByTokenRequest) (*dto.RoomResponse, error) {
	room, err := s.repo.FindRoomByInviteToken(ctx, req.InviteToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return s.joinRoom(ctx, room, req.UserUUID)
}

func (s *roomService) joinRoom(ctx context.Context, room *models.Room, userUUID string) (*dto.RoomResponse, error) {
	userID, err := s.resolveUserID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	isMember, err := s.repo.IsMember(ctx, room.ID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		if err := s.repo.AddMember(ctx, &models.RoomMember{
			RoomID: room.ID,
			UserID: userID,
			Role:   models.RoleMember,
		}); err != nil {
			return nil, err
		}
	}

	if err := s.stateRepo.AddOnlineMember(ctx, room.UUID.String(), userUUID); err != nil {
		return nil, err
	}

	return s.toResponse(ctx, room, userUUID, false)
}

func (s *roomService) LeaveRoom(ctx context.Context, roomUUID, userUUID string) error {
	room, err := s.repo.FindRoomByUUID(ctx, mustParseUUID(roomUUID))
	if err != nil {
		return ErrRoomNotFound
	}

	userID, err := s.resolveUserID(ctx, userUUID)
	if err != nil {
		return err
	}

	if room.HostID == userID {
		return errors.New("host cannot leave room, transfer ownership or delete room instead")
	}

	if err := s.stateRepo.RemoveOnlineMember(ctx, roomUUID, userUUID); err != nil {
		return err
	}

	return s.repo.RemoveMember(ctx, room.ID, userID)
}

func (s *roomService) ListPublicRooms(ctx context.Context, page, pageSize int) ([]dto.RoomResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rooms, total, err := s.repo.ListPublicRooms(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		resp, err := s.toResponse(ctx, &room, "", false)
		if err != nil {
			continue
		}
		result = append(result, *resp)
	}

	return result, total, nil
}

func (s *roomService) isRequesterHost(ctx context.Context, room *models.Room, requesterUUID string) (bool, error) {
	if requesterUUID == "" {
		return false, nil
	}
	userID, err := s.resolveUserID(ctx, requesterUUID)
	if err != nil {
		return false, nil // requester tidak valid, anggap bukan host, jangan block response
	}
	return room.HostID == userID, nil
}

func (s *roomService) toResponse(ctx context.Context, room *models.Room, requesterUUID string, isHost bool) (*dto.RoomResponse, error) {
	memberCount, err := s.repo.CountMembers(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	onlineCount, err := s.stateRepo.CountOnlineMembers(ctx, room.UUID.String())
	if err != nil {
		onlineCount = 0
	}

	hostResp, err := s.authClient.GetUserByID(ctx, &authv1.GetUserByIDRequest{Id: room.HostID})
	if err != nil {
		return nil, errors.New("failed to resolve host information")
	}

	resp := &dto.RoomResponse{
		UUID:        room.UUID.String(),
		Name:        room.Name,
		Visibility:  room.Visibility,
		HostUUID:    hostResp.Uuid,
		MemberCount: int(memberCount),
		OnlineCount: int(onlineCount),
	}

	if isHost {
		resp.InviteCode = room.InviteCode
		resp.InviteToken = room.InviteToken
	}

	return resp, nil
}

func generateInviteCode() (string, error) {
	const chars = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		code[i] = chars[n.Int64()]
	}
	return string(code), nil
}

func generateInviteToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *roomService) JoinRoom(ctx context.Context, req dto.JoinRoomRequest) (*dto.RoomResponse, error) {
	room, err := s.repo.FindRoomByUUID(ctx, mustParseUUID(req.RoomUUID))
	if err != nil {
		return nil, ErrRoomNotFound
	}

	if room.Visibility != models.VisibilityPublic {
		return nil, errors.New("this room requires an invite code or link")
	}

	return s.joinRoom(ctx, room, req.UserUUID)
}

func (s *roomService) ListMyRooms(ctx context.Context, userUUID string, page, pageSize int) ([]dto.RoomResponse, int64, error) {
	userID, err := s.resolveUserID(ctx, userUUID)
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rooms, total, err := s.repo.ListRoomsByMemberUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		isHost := room.HostID == userID
		resp, err := s.toResponse(ctx, &room, userUUID, isHost)
		if err != nil {
			continue
		}
		result = append(result, *resp)
	}

	return result, total, nil
}

func (s *roomService) GetRoomInternal(ctx context.Context, roomUUID string) (int64, int64, error) {
	room, err := s.repo.FindRoomByUUID(ctx, mustParseUUID(roomUUID))
	if err != nil {
		return 0, 0, ErrRoomNotFound
	}
	return room.ID, room.HostID, nil
}

func (s *roomService) CheckMembership(ctx context.Context, roomUUID, userUUID string) (isMember bool, isHost bool, err error) {
	room, err := s.repo.FindRoomByUUID(ctx, mustParseUUID(roomUUID))
	if err != nil {
		return false, false, ErrRoomNotFound
	}

	userID, err := s.resolveUserID(ctx, userUUID)
	if err != nil {
		return false, false, ErrUserNotFound
	}

	isMember, err = s.repo.IsMember(ctx, room.ID, userID)
	if err != nil {
		return false, false, err
	}

	isHost = room.HostID == userID
	return isMember, isHost, nil
}

type PlaybackState struct {
	CurrentTrackID  string
	PositionSeconds int
	IsPlaying       bool
	UpdatedAt       string
}

func (s *roomService) UpdatePlayback(ctx context.Context, roomUUID, requesterUUID, currentTrackID string, positionSeconds int, isPlaying bool) (*PlaybackState, error) {
	// hanya host yang boleh mengubah playback
	isMember, isHost, err := s.CheckMembership(ctx, roomUUID, requesterUUID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotAuthorized
	}
	if !isHost {
		return nil, ErrNotAuthorized
	}

	state := repository.RoomState{
		CurrentTrackID:  currentTrackID,
		PositionSeconds: positionSeconds,
		IsPlaying:       isPlaying,
	}
	if err := s.stateRepo.UpdateState(ctx, roomUUID, state); err != nil {
		return nil, err
	}

	updated, err := s.stateRepo.GetState(ctx, roomUUID)
	if err != nil {
		return nil, err
	}
	return toPlaybackState(updated), nil
}

func (s *roomService) GetPlayback(ctx context.Context, roomUUID, requesterUUID string) (*PlaybackState, error) {
	// member mana pun boleh membaca state
	isMember, _, err := s.CheckMembership(ctx, roomUUID, requesterUUID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotAuthorized
	}

	state, err := s.stateRepo.GetState(ctx, roomUUID)
	if err != nil {
		return nil, err
	}
	return toPlaybackState(state), nil
}

func toPlaybackState(s *repository.RoomState) *PlaybackState {
	if s == nil {
		return &PlaybackState{} // state belum pernah di-set
	}
	return &PlaybackState{
		CurrentTrackID:  s.CurrentTrackID,
		PositionSeconds: s.PositionSeconds,
		IsPlaying:       s.IsPlaying,
		UpdatedAt:       s.UpdatedAt.Format(time.RFC3339),
	}
}
