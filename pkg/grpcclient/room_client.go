package grpcclient

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	roomv1 "listenly-backend/gen/go/room/v1"
)

func NewRoomServiceClient(addr string) (roomv1.RoomServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return roomv1.NewRoomServiceClient(conn), nil
}
