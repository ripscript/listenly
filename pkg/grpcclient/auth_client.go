package grpcclient

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "listenly-backend/gen/go/auth/v1"
)

func NewAuthServiceClient(addr string) (authv1.AuthServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return authv1.NewAuthServiceClient(conn), nil
}
