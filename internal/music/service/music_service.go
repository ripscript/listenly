package service

import (
	"context"
	"errors"
	"log/slog"

	authv1 "listenly-backend/gen/go/auth/v1"
	mediav1 "listenly-backend/gen/go/media/v1"
	roomv1 "listenly-backend/gen/go/room/v1"
	"listenly-backend/internal/music/dto"
	"listenly-backend/internal/music/models"
	"listenly-backend/internal/music/repository"
)

var (
	ErrTrackNotFound        = errors.New("track not found")
	ErrQueueItemNotFound    = errors.New("queue item not found")
	ErrTrackNotYetSupported = errors.New("track not found in system; media extraction not yet supported")
	ErrNotAuthorized        = errors.New("not authorized to remove this item")
)

type MusicService interface {
	SearchTracks(ctx context.Context, req dto.SearchTracksRequest) ([]dto.TrackResponse, int64, error)
	GetTrack(ctx context.Context, trackUUID string) (*dto.TrackResponse, error)
	RequestTrack(ctx context.Context, req dto.RequestTrackRequest) (*dto.QueueItemResponse, error)
	GetQueue(ctx context.Context, roomUUID, requesterUUID string) ([]dto.QueueItemResponse, error)
	RemoveFromQueue(ctx context.Context, req dto.RemoveFromQueueRequest) error
	MarkAsPlayed(ctx context.Context, queueItemUUID string) error
}

type musicService struct {
	repo        repository.MusicRepository
	searchRepo  repository.TrackSearchRepository
	authClient  authv1.AuthServiceClient
	roomClient  roomv1.RoomServiceClient
	mediaClient mediav1.MediaServiceClient
}

var ErrNotRoomMember = errors.New("you are not a member of this room")

func NewMusicService(
	repo repository.MusicRepository,
	searchRepo repository.TrackSearchRepository,
	authClient authv1.AuthServiceClient,
	roomClient roomv1.RoomServiceClient,
	mediaClient mediav1.MediaServiceClient,
) MusicService {
	return &musicService{
		repo:        repo,
		searchRepo:  searchRepo,
		authClient:  authClient,
		roomClient:  roomClient,
		mediaClient: mediaClient,
	}
}

func (s *musicService) SearchTracks(ctx context.Context, req dto.SearchTracksRequest) ([]dto.TrackResponse, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize

	uuids, total, err := s.searchRepo.Search(ctx, req.Query, req.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(uuids) == 0 {
		return []dto.TrackResponse{}, 0, nil
	}

	tracks, err := s.repo.FindTracksByUUIDs(ctx, uuids)
	if err != nil {
		return nil, 0, err
	}

	// Urutkan hasil sesuai urutan relevansi dari Elasticsearch (bukan urutan dari Postgres)
	trackMap := make(map[string]models.Track, len(tracks))
	for _, t := range tracks {
		trackMap[t.UUID.String()] = t
	}

	result := make([]dto.TrackResponse, 0, len(uuids))
	for _, u := range uuids {
		if t, ok := trackMap[u]; ok {
			result = append(result, toTrackResponse(&t))
		}
	}

	return result, total, nil
}

func (s *musicService) GetTrack(ctx context.Context, trackUUID string) (*dto.TrackResponse, error) {
	track, err := s.repo.FindTrackByUUID(ctx, mustParseUUID(trackUUID))
	if err != nil {
		return nil, ErrTrackNotFound
	}
	resp := toTrackResponse(track)
	return &resp, nil
}

func (s *musicService) RequestTrack(ctx context.Context, req dto.RequestTrackRequest) (*dto.QueueItemResponse, error) {
	roomInternal, err := s.roomClient.GetRoomInternal(ctx, &roomv1.GetRoomInternalRequest{RoomUuid: req.RoomUUID})
	if err != nil {
		return nil, errors.New("room not found")
	}

	membership, err := s.roomClient.CheckMembership(ctx, &roomv1.CheckMembershipRequest{
		RoomUuid: req.RoomUUID,
		UserUuid: req.RequestedByUUID,
	})
	if err != nil || !membership.IsMember {
		return nil, ErrNotRoomMember
	}

	userResp, err := s.authClient.GetUserByUUID(ctx, &authv1.GetUserByUUIDRequest{Uuid: req.RequestedByUUID})
	if err != nil {
		return nil, errors.New("user not found")
	}

	track, err := s.repo.FindTrackByYoutubeID(ctx, req.YoutubeVideoID)
	if err != nil {
		// Track belum ada di sistem -> fetch dari Media Service sekarang
		track, err = s.fetchAndCreateTrack(ctx, req.YoutubeVideoID)
		if err != nil {
			return nil, err
		}
	}

	position, err := s.repo.CountQueueByRoomID(ctx, roomInternal.Id)
	if err != nil {
		return nil, err
	}

	item := &models.QueueItem{
		RoomID:      roomInternal.Id,
		RoomHostID:  roomInternal.HostId,
		TrackID:     track.ID,
		RequestedBy: userResp.Id,
		Status:      models.StatusReady,
		Position:    int(position),
	}

	if err := s.repo.CreateQueueItem(ctx, item); err != nil {
		return nil, err
	}

	item.Track = *track
	resp := toQueueItemResponse(item, req.RequestedByUUID)
	return &resp, nil
}

func (s *musicService) GetQueue(ctx context.Context, roomUUID, requesterUUID string) ([]dto.QueueItemResponse, error) {
	roomInternal, err := s.roomClient.GetRoomInternal(ctx, &roomv1.GetRoomInternalRequest{RoomUuid: roomUUID})
	if err != nil {
		return nil, errors.New("room not found")
	}

	membership, err := s.roomClient.CheckMembership(ctx, &roomv1.CheckMembershipRequest{
		RoomUuid: roomUUID,
		UserUuid: requesterUUID,
	})
	if err != nil || !membership.IsMember {
		return nil, ErrNotRoomMember
	}

	items, err := s.repo.ListQueueByRoomID(ctx, roomInternal.Id)
	if err != nil {
		return nil, err
	}

	result := make([]dto.QueueItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, toQueueItemResponse(&item, ""))
	}
	return result, nil
}

func (s *musicService) RemoveFromQueue(ctx context.Context, req dto.RemoveFromQueueRequest) error {
	item, err := s.repo.FindQueueItemByUUID(ctx, mustParseUUID(req.QueueItemUUID))
	if err != nil {
		return ErrQueueItemNotFound
	}

	userResp, err := s.authClient.GetUserByUUID(ctx, &authv1.GetUserByUUIDRequest{Uuid: req.RequesterUUID})
	if err != nil {
		return errors.New("user not found")
	}

	isOwner := item.RequestedBy == userResp.Id
	isHost := item.RoomHostID == userResp.Id

	if !isOwner && !isHost {
		return ErrNotAuthorized
	}

	return s.repo.DeleteQueueItem(ctx, item.ID)
}

func (s *musicService) MarkAsPlayed(ctx context.Context, queueItemUUID string) error {
	item, err := s.repo.FindQueueItemByUUID(ctx, mustParseUUID(queueItemUUID))
	if err != nil {
		return ErrQueueItemNotFound
	}
	return s.repo.UpdateQueueItemStatus(ctx, item.ID, models.StatusPlayed)
}

func toTrackResponse(t *models.Track) dto.TrackResponse {
	return dto.TrackResponse{
		UUID:            t.UUID.String(),
		YoutubeVideoID:  t.YoutubeVideoID,
		Title:           t.Title,
		Artist:          t.Artist,
		DurationSeconds: t.DurationSeconds,
		ThumbnailURL:    t.ThumbnailURL,
	}
}

func toQueueItemResponse(item *models.QueueItem, requestedByUUID string) dto.QueueItemResponse {
	return dto.QueueItemResponse{
		UUID:            item.UUID.String(),
		Track:           toTrackResponse(&item.Track),
		RequestedByUUID: requestedByUUID,
		Status:          item.Status,
		Position:        item.Position,
		CreatedAt:       item.CreatedAt,
	}
}

func (s *musicService) fetchAndCreateTrack(ctx context.Context, youtubeVideoID string) (*models.Track, error) {
	youtubeURL := "https://www.youtube.com/watch?v=" + youtubeVideoID

	info, err := s.mediaClient.GetMediaInfo(ctx, &mediav1.GetMediaInfoRequest{YoutubeUrl: youtubeURL})
	if err != nil {
		return nil, errors.New("failed to fetch media info: video may be unavailable")
	}

	track := &models.Track{
		YoutubeVideoID:  info.VideoId,
		Title:           info.Title,
		Artist:          info.Channel,
		DurationSeconds: int(info.DurationSeconds),
		ThumbnailURL:    info.ThumbnailUrl,
	}

	if err := s.repo.CreateTrack(ctx, track); err != nil {
		return nil, err
	}

	// Index ke Elasticsearch langsung saat create (real-time, BUKAN startup-based —
	// ini sekaligus menuntaskan utang teknis 3.11 di listenly-progress.md)
	if err := s.searchRepo.IndexTrack(ctx, track); err != nil {
		slog.Warn("failed to index track to elasticsearch", "track_id", track.ID, "error", err)
	}

	return track, nil
}
