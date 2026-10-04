package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc/status"

	roomv1 "listenly-backend/gen/go/room/v1"
	"listenly-backend/internal/gateway/dto"
	"listenly-backend/pkg/ctxutil"
	"listenly-backend/pkg/response"
)

type RoomHandler struct {
	roomClient roomv1.RoomServiceClient
	webBaseURL string
}

func NewRoomHandler(roomClient roomv1.RoomServiceClient, webBaseURL string) *RoomHandler {
	return &RoomHandler{
		roomClient: roomClient,
		webBaseURL: webBaseURL,
	}
}

func (h *RoomHandler) CreateRoom(c *echo.Context) error {
	var req dto.CreateRoomHTTPRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.ValidationErrorResponse(c, err)
	}

	hostUUID := ctxutil.GetUserUUID(c)

	visibility := roomv1.RoomVisibility_ROOM_VISIBILITY_PUBLIC
	if req.Visibility == "private" {
		visibility = roomv1.RoomVisibility_ROOM_VISIBILITY_PRIVATE
	}

	result, err := h.roomClient.CreateRoom(c.Request().Context(), &roomv1.CreateRoomRequest{
		HostUuid:   hostUUID,
		Name:       req.Name,
		Visibility: visibility,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusBadRequest, st.Message(), nil)
	}

	return response.Success(c, http.StatusCreated, "room created", h.toRoomHTTPResponse(result))
}

func (h *RoomHandler) GetRoom(c *echo.Context) error {
	roomUUID := c.Param("uuid")
	requesterUUID := ctxutil.GetUserUUID(c)

	result, err := h.roomClient.GetRoom(c.Request().Context(), &roomv1.GetRoomRequest{
		RoomUuid:      roomUUID,
		RequesterUuid: requesterUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusNotFound, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "room found", h.toRoomHTTPResponse(result))
}

func (h *RoomHandler) LeaveRoom(c *echo.Context) error {
	roomUUID := c.Param("uuid")
	userUUID := ctxutil.GetUserUUID(c)

	_, err := h.roomClient.LeaveRoom(c.Request().Context(), &roomv1.LeaveRoomRequest{
		RoomUuid: roomUUID,
		UserUuid: userUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusBadRequest, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "left room", nil)
}

func (h *RoomHandler) ListPublicRooms(c *echo.Context) error {
	page, pageSize := parsePagination(c)

	result, err := h.roomClient.ListPublicRooms(c.Request().Context(), &roomv1.ListPublicRoomsRequest{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusInternalServerError, st.Message(), nil)
	}

	rooms := make([]dto.RoomHTTPResponse, 0, len(result.Rooms))
	for _, r := range result.Rooms {
		rooms = append(rooms, h.toRoomHTTPResponse(r))
	}

	return response.Success(c, http.StatusOK, "public rooms", dto.ListPublicRoomsHTTPResponse{
		Rooms:      rooms,
		TotalItems: int(result.TotalItems),
	})
}

func (h *RoomHandler) toRoomHTTPResponse(r *roomv1.RoomResponse) dto.RoomHTTPResponse {
	resp := dto.RoomHTTPResponse{
		UUID:        r.Uuid,
		Name:        r.Name,
		Visibility:  visibilityToStringHTTP(r.Visibility),
		HostUUID:    r.HostUuid,
		InviteCode:  r.InviteCode,
		MemberCount: int(r.MemberCount),
		OnlineCount: int(r.OnlineCount),
	}

	if r.InviteToken != nil {
		link := h.webBaseURL + "/join/" + *r.InviteToken
		resp.InviteLink = &link
	}

	return resp
}

func visibilityToStringHTTP(v roomv1.RoomVisibility) string {
	if v == roomv1.RoomVisibility_ROOM_VISIBILITY_PRIVATE {
		return "private"
	}
	return "public"
}

func (h *RoomHandler) JoinRoom(c *echo.Context) error {
	roomUUID := c.Param("uuid")
	userUUID := ctxutil.GetUserUUID(c)

	result, err := h.roomClient.JoinRoom(c.Request().Context(), &roomv1.JoinRoomRequest{
		RoomUuid: roomUUID,
		UserUuid: userUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusBadRequest, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "joined room", h.toRoomHTTPResponse(result))
}

func (h *RoomHandler) JoinPrivateRoom(c *echo.Context) error {
	var req dto.JoinPrivateRoomHTTPRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}

	userUUID := ctxutil.GetUserUUID(c)
	var result *roomv1.RoomResponse
	var err error

	switch {
	case req.InviteCode != nil:
		result, err = h.roomClient.JoinRoomByCode(c.Request().Context(), &roomv1.JoinRoomByCodeRequest{
			UserUuid:   userUUID,
			InviteCode: *req.InviteCode,
		})
	case req.InviteToken != nil:
		result, err = h.roomClient.JoinRoomByToken(c.Request().Context(), &roomv1.JoinRoomByTokenRequest{
			UserUuid:    userUUID,
			InviteToken: *req.InviteToken,
		})
	default:
		return response.Error(c, http.StatusBadRequest, "invite_code or invite_token is required", nil)
	}

	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusBadRequest, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "joined room", h.toRoomHTTPResponse(result))
}

func (h *RoomHandler) ListMyRooms(c *echo.Context) error {
	userUUID := ctxutil.GetUserUUID(c)

	page, pageSize := parsePagination(c)

	result, err := h.roomClient.ListMyRooms(c.Request().Context(), &roomv1.ListMyRoomsRequest{
		UserUuid: userUUID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusInternalServerError, st.Message(), nil)
	}

	rooms := make([]dto.RoomHTTPResponse, 0, len(result.Rooms))
	for _, r := range result.Rooms {
		rooms = append(rooms, h.toRoomHTTPResponse(r))
	}

	return response.Success(c, http.StatusOK, "my rooms", dto.ListPublicRoomsHTTPResponse{
		Rooms:      rooms,
		TotalItems: int(result.TotalItems),
	})
}

func (h *RoomHandler) UpdatePlayback(c *echo.Context) error {
	roomUUID := c.Param("uuid")
	requesterUUID := ctxutil.GetUserUUID(c)

	var req dto.UpdatePlaybackHTTPRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}

	result, err := h.roomClient.UpdatePlayback(c.Request().Context(), &roomv1.UpdatePlaybackRequest{
		RoomUuid:        roomUUID,
		RequesterUuid:   requesterUUID,
		CurrentTrackId:  req.CurrentTrackID,
		PositionSeconds: int32(req.PositionSeconds),
		IsPlaying:       req.IsPlaying,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusForbidden, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "playback updated", toPlaybackHTTPResponse(result))
}

func (h *RoomHandler) GetPlayback(c *echo.Context) error {
	roomUUID := c.Param("uuid")
	requesterUUID := ctxutil.GetUserUUID(c)

	result, err := h.roomClient.GetPlayback(c.Request().Context(), &roomv1.GetPlaybackRequest{
		RoomUuid:      roomUUID,
		RequesterUuid: requesterUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusForbidden, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "playback state", toPlaybackHTTPResponse(result))
}

func toPlaybackHTTPResponse(s *roomv1.PlaybackStateResponse) dto.PlaybackStateHTTPResponse {
	return dto.PlaybackStateHTTPResponse{
		CurrentTrackID:  s.CurrentTrackId,
		PositionSeconds: int(s.PositionSeconds),
		IsPlaying:       s.IsPlaying,
		UpdatedAt:       s.UpdatedAt,
	}
}
