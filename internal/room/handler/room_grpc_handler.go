package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	roomv1 "listenly-backend/gen/go/room/v1"
	"listenly-backend/internal/room/dto"
	"listenly-backend/internal/room/service"
)

type RoomGRPCHandler struct {
	roomv1.UnimplementedRoomServiceServer
	service service.RoomService
}

func NewRoomGRPCHandler(service service.RoomService) *RoomGRPCHandler {
	return &RoomGRPCHandler{service: service}
}

func (h *RoomGRPCHandler) CreateRoom(ctx context.Context, req *roomv1.CreateRoomRequest) (*roomv1.RoomResponse, error) {
	result, err := h.service.CreateRoom(ctx, dto.CreateRoomRequest{
		HostUUID:   req.GetHostUuid(),
		Name:       req.GetName(),
		Visibility: visibilityToString(req.GetVisibility()),
	})
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoRoomResponse(result), nil
}

func (h *RoomGRPCHandler) GetRoom(ctx context.Context, req *roomv1.GetRoomRequest) (*roomv1.RoomResponse, error) {
	result, err := h.service.GetRoom(ctx, req.GetRoomUuid(), req.GetRequesterUuid())
	if err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoRoomResponse(result), nil
}

func (h *RoomGRPCHandler) JoinRoomByCode(ctx context.Context, req *roomv1.JoinRoomByCodeRequest) (*roomv1.RoomResponse, error) {
	result, err := h.service.JoinRoomByCode(ctx, dto.JoinRoomByCodeRequest{
		UserUUID:   req.GetUserUuid(),
		InviteCode: req.GetInviteCode(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoRoomResponse(result), nil
}

func (h *RoomGRPCHandler) JoinRoomByToken(ctx context.Context, req *roomv1.JoinRoomByTokenRequest) (*roomv1.RoomResponse, error) {
	result, err := h.service.JoinRoomByToken(ctx, dto.JoinRoomByTokenRequest{
		UserUUID:    req.GetUserUuid(),
		InviteToken: req.GetInviteToken(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoRoomResponse(result), nil
}

func (h *RoomGRPCHandler) LeaveRoom(ctx context.Context, req *roomv1.LeaveRoomRequest) (*roomv1.LeaveRoomResponse, error) {
	if err := h.service.LeaveRoom(ctx, req.GetRoomUuid(), req.GetUserUuid()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &roomv1.LeaveRoomResponse{Success: true}, nil
}

func (h *RoomGRPCHandler) ListPublicRooms(ctx context.Context, req *roomv1.ListPublicRoomsRequest) (*roomv1.ListPublicRoomsResponse, error) {
	rooms, total, err := h.service.ListPublicRooms(ctx, int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoRooms := make([]*roomv1.RoomResponse, 0, len(rooms))
	for _, r := range rooms {
		protoRooms = append(protoRooms, toProtoRoomResponse(&r))
	}

	return &roomv1.ListPublicRoomsResponse{
		Rooms:      protoRooms,
		TotalItems: int32(total),
	}, nil
}

func toProtoRoomResponse(r *dto.RoomResponse) *roomv1.RoomResponse {
	return &roomv1.RoomResponse{
		Uuid:        r.UUID,
		Name:        r.Name,
		Visibility:  stringToVisibility(r.Visibility),
		HostUuid:    r.HostUUID,
		InviteCode:  r.InviteCode,
		InviteToken: r.InviteToken,
		MemberCount: int32(r.MemberCount),
		OnlineCount: int32(r.OnlineCount),
	}
}

func visibilityToString(v roomv1.RoomVisibility) string {
	if v == roomv1.RoomVisibility_ROOM_VISIBILITY_PRIVATE {
		return "private"
	}
	return "public"
}

func stringToVisibility(s string) roomv1.RoomVisibility {
	if s == "private" {
		return roomv1.RoomVisibility_ROOM_VISIBILITY_PRIVATE
	}
	return roomv1.RoomVisibility_ROOM_VISIBILITY_PUBLIC
}

func (h *RoomGRPCHandler) JoinRoom(ctx context.Context, req *roomv1.JoinRoomRequest) (*roomv1.RoomResponse, error) {
	result, err := h.service.JoinRoom(ctx, dto.JoinRoomRequest{
		RoomUUID: req.GetRoomUuid(),
		UserUUID: req.GetUserUuid(),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoRoomResponse(result), nil
}

func (h *RoomGRPCHandler) ListMyRooms(ctx context.Context, req *roomv1.ListMyRoomsRequest) (*roomv1.ListPublicRoomsResponse, error) {
	rooms, total, err := h.service.ListMyRooms(ctx, req.GetUserUuid(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoRooms := make([]*roomv1.RoomResponse, 0, len(rooms))
	for _, r := range rooms {
		protoRooms = append(protoRooms, toProtoRoomResponse(&r))
	}

	return &roomv1.ListPublicRoomsResponse{
		Rooms:      protoRooms,
		TotalItems: int32(total),
	}, nil
}

func (h *RoomGRPCHandler) GetRoomInternal(ctx context.Context, req *roomv1.GetRoomInternalRequest) (*roomv1.RoomInternalResponse, error) {
	id, hostID, err := h.service.GetRoomInternal(ctx, req.GetRoomUuid())
	if err != nil {
		return nil, status.Error(codes.NotFound, "room not found")
	}
	return &roomv1.RoomInternalResponse{
		Id:     id,
		Uuid:   req.GetRoomUuid(),
		HostId: hostID,
	}, nil
}

func (h *RoomGRPCHandler) CheckMembership(ctx context.Context, req *roomv1.CheckMembershipRequest) (*roomv1.CheckMembershipResponse, error) {
	isMember, isHost, err := h.service.CheckMembership(ctx, req.GetRoomUuid(), req.GetUserUuid())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &roomv1.CheckMembershipResponse{
		IsMember: isMember,
		IsHost:   isHost,
	}, nil
}
