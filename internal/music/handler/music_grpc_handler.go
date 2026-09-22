package handler

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	musicv1 "listenly-backend/gen/go/music/v1"
	"listenly-backend/internal/music/dto"
	"listenly-backend/internal/music/service"
)

type MusicGRPCHandler struct {
	musicv1.UnimplementedMusicServiceServer
	service service.MusicService
}

func NewMusicGRPCHandler(service service.MusicService) *MusicGRPCHandler {
	return &MusicGRPCHandler{service: service}
}

func (h *MusicGRPCHandler) SearchTracks(ctx context.Context, req *musicv1.SearchTracksRequest) (*musicv1.SearchTracksResponse, error) {
	tracks, total, err := h.service.SearchTracks(ctx, dto.SearchTracksRequest{
		Query:    req.GetQuery(),
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoTracks := make([]*musicv1.TrackResponse, 0, len(tracks))
	for _, t := range tracks {
		protoTracks = append(protoTracks, toProtoTrackResponse(&t))
	}

	return &musicv1.SearchTracksResponse{
		Tracks:     protoTracks,
		TotalItems: int32(total),
	}, nil
}

func (h *MusicGRPCHandler) GetTrack(ctx context.Context, req *musicv1.GetTrackRequest) (*musicv1.TrackResponse, error) {
	track, err := h.service.GetTrack(ctx, req.GetTrackUuid())
	if err != nil {
		if errors.Is(err, service.ErrTrackNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoTrackResponse(track), nil
}

func (h *MusicGRPCHandler) RequestTrack(ctx context.Context, req *musicv1.RequestTrackRequest) (*musicv1.QueueItemResponse, error) {
	result, err := h.service.RequestTrack(ctx, dto.RequestTrackRequest{
		RoomUUID:        req.GetRoomUuid(),
		RequestedByUUID: req.GetRequestedByUuid(),
		YoutubeVideoID:  req.GetYoutubeVideoId(),
	})
	if err != nil {
		if errors.Is(err, service.ErrTrackNotYetSupported) {
			return nil, status.Error(codes.Unimplemented, err.Error())
		}

		if errors.Is(err, service.ErrNotRoomMember) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoQueueItemResponse(result), nil
}

func (h *MusicGRPCHandler) GetQueue(ctx context.Context, req *musicv1.GetQueueRequest) (*musicv1.GetQueueResponse, error) {
	items, err := h.service.GetQueue(ctx, req.GetRoomUuid(), req.GetRequesterUuid())
	if err != nil {
		if errors.Is(err, service.ErrNotRoomMember) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoItems := make([]*musicv1.QueueItemResponse, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, toProtoQueueItemResponse(&item))
	}

	return &musicv1.GetQueueResponse{Items: protoItems}, nil
}

func (h *MusicGRPCHandler) RemoveFromQueue(ctx context.Context, req *musicv1.RemoveFromQueueRequest) (*musicv1.RemoveFromQueueResponse, error) {
	err := h.service.RemoveFromQueue(ctx, dto.RemoveFromQueueRequest{
		QueueItemUUID: req.GetQueueItemUuid(),
		RequesterUUID: req.GetRequesterUuid(),
	})
	if err != nil {
		if errors.Is(err, service.ErrNotAuthorized) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if errors.Is(err, service.ErrQueueItemNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &musicv1.RemoveFromQueueResponse{Success: true}, nil
}

func (h *MusicGRPCHandler) MarkAsPlayed(ctx context.Context, req *musicv1.MarkAsPlayedRequest) (*musicv1.MarkAsPlayedResponse, error) {
	if err := h.service.MarkAsPlayed(ctx, req.GetQueueItemUuid()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &musicv1.MarkAsPlayedResponse{Success: true}, nil
}

func toProtoTrackResponse(t *dto.TrackResponse) *musicv1.TrackResponse {
	return &musicv1.TrackResponse{
		Uuid:            t.UUID,
		YoutubeVideoId:  t.YoutubeVideoID,
		Title:           t.Title,
		Artist:          t.Artist,
		DurationSeconds: int32(t.DurationSeconds),
		ThumbnailUrl:    t.ThumbnailURL,
	}
}

func toProtoQueueItemResponse(item *dto.QueueItemResponse) *musicv1.QueueItemResponse {
	return &musicv1.QueueItemResponse{
		Uuid:            item.UUID,
		Track:           toProtoTrackResponse(&item.Track),
		RequestedByUuid: item.RequestedByUUID,
		Status:          item.Status,
		Position:        int32(item.Position),
		CreatedAt:       item.CreatedAt.Format(time.RFC3339),
	}
}
