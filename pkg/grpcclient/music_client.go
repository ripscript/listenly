package grpcclient

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	musicv1 "listenly-backend/gen/go/music/v1"
)

func NewMusicServiceClient(addr string) (musicv1.MusicServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return musicv1.NewMusicServiceClient(conn), nil
}
