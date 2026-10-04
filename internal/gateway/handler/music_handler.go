package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc/status"

	musicv1 "listenly-backend/gen/go/music/v1"
	"listenly-backend/internal/gateway/dto"
	"listenly-backend/pkg/ctxutil"
	"listenly-backend/pkg/response"
)

type MusicHandler struct {
	musicClient musicv1.MusicServiceClient
}

func NewMusicHandler(musicClient musicv1.MusicServiceClient) *MusicHandler {
	return &MusicHandler{musicClient: musicClient}
}

func (h *MusicHandler) SearchTracks(c *echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return response.Error(c, http.StatusBadRequest, "query parameter 'q' is required", nil)
	}

	page, pageSize := parsePagination(c)

	result, err := h.musicClient.SearchTracks(c.Request().Context(), &musicv1.SearchTracksRequest{
		Query:    query,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusInternalServerError, st.Message(), nil)
	}

	tracks := make([]dto.TrackHTTPResponse, 0, len(result.Tracks))
	for _, t := range result.Tracks {
		tracks = append(tracks, toTrackHTTPResponse(t))
	}

	return response.Success(c, http.StatusOK, "search results", dto.SearchTracksHTTPResponse{
		Tracks:     tracks,
		TotalItems: int(result.TotalItems),
	})
}

func (h *MusicHandler) RequestTrack(c *echo.Context) error {
	roomUUID := c.Param("roomUuid")

	var req dto.RequestTrackHTTPRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.ValidationErrorResponse(c, err)
	}

	requesterUUID := ctxutil.GetUserUUID(c)

	result, err := h.musicClient.RequestTrack(c.Request().Context(), &musicv1.RequestTrackRequest{
		RoomUuid:        roomUUID,
		RequestedByUuid: requesterUUID,
		YoutubeVideoId:  req.YoutubeVideoID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusBadRequest, st.Message(), nil)
	}

	return response.Success(c, http.StatusCreated, "track requested", toQueueItemHTTPResponse(result))
}

func (h *MusicHandler) GetQueue(c *echo.Context) error {
	roomUUID := c.Param("roomUuid")
	requesterUUID := ctxutil.GetUserUUID(c)

	result, err := h.musicClient.GetQueue(c.Request().Context(), &musicv1.GetQueueRequest{
		RoomUuid:      roomUUID,
		RequesterUuid: requesterUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusForbidden, st.Message(), nil)
	}

	items := make([]dto.QueueItemHTTPResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toQueueItemHTTPResponse(item))
	}

	return response.Success(c, http.StatusOK, "queue", items)
}

func (h *MusicHandler) RemoveFromQueue(c *echo.Context) error {
	queueItemUUID := c.Param("uuid")
	requesterUUID := ctxutil.GetUserUUID(c)

	_, err := h.musicClient.RemoveFromQueue(c.Request().Context(), &musicv1.RemoveFromQueueRequest{
		QueueItemUuid: queueItemUUID,
		RequesterUuid: requesterUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusBadRequest, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "removed from queue", nil)
}

func toTrackHTTPResponse(t *musicv1.TrackResponse) dto.TrackHTTPResponse {
	return dto.TrackHTTPResponse{
		UUID:            t.Uuid,
		YoutubeVideoID:  t.YoutubeVideoId,
		Title:           t.Title,
		Artist:          t.Artist,
		DurationSeconds: int(t.DurationSeconds),
		ThumbnailURL:    t.ThumbnailUrl,
	}
}

func toQueueItemHTTPResponse(item *musicv1.QueueItemResponse) dto.QueueItemHTTPResponse {
	return dto.QueueItemHTTPResponse{
		UUID:            item.Uuid,
		Track:           toTrackHTTPResponse(item.Track),
		RequestedByUUID: item.RequestedByUuid,
		Status:          item.Status,
		Position:        int(item.Position),
		CreatedAt:       item.CreatedAt,
	}
}

func (h *MusicHandler) MarkAsPlayed(c *echo.Context) error {
	queueItemUUID := c.Param("uuid")

	_, err := h.musicClient.MarkAsPlayed(c.Request().Context(), &musicv1.MarkAsPlayedRequest{
		QueueItemUuid: queueItemUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusBadRequest, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "marked as played", nil)
}

func (h *MusicHandler) GetTrack(c *echo.Context) error {
	trackUUID := c.Param("uuid")

	result, err := h.musicClient.GetTrack(c.Request().Context(), &musicv1.GetTrackRequest{
		TrackUuid: trackUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusNotFound, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "track found", toTrackHTTPResponse(result))
}
