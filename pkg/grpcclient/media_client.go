package grpcclient

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	mediav1 "listenly-backend/gen/go/media/v1"
)

func NewMediaServiceClient(addr string) (mediav1.MediaServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return mediav1.NewMediaServiceClient(conn), nil
}
