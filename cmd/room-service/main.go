package main

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"listenly-backend/config"
	roomv1 "listenly-backend/gen/go/room/v1"
	"listenly-backend/internal/room/handler"
	"listenly-backend/internal/room/repository"
	"listenly-backend/internal/room/service"
	"listenly-backend/pkg/database"
	"listenly-backend/pkg/grpcclient"
	"listenly-backend/pkg/grpcserver"
	"listenly-backend/pkg/redis"
)

func main() {
	config.LoadEnv("cmd/room-service/.env")
	cfg := config.LoadRoomConfig()

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		return
	}

	rdb, err := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Error("failed to connect redis", "error", err)
		return
	}

	authClient, err := grpcclient.NewAuthServiceClient(cfg.AuthServiceAddr)
	if err != nil {
		slog.Error("failed to connect auth-service", "error", err)
		return
	}

	roomRepo := repository.NewRoomRepository(db)
	roomStateRepo := repository.NewRoomStateRepository(rdb)
	roomSvc := service.NewRoomService(roomRepo, roomStateRepo, authClient)
	roomHandler := handler.NewRoomGRPCHandler(roomSvc)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.RecoveryInterceptor),
	)
	roomv1.RegisterRoomServiceServer(grpcServer, roomHandler)
	reflection.Register(grpcServer)

	slog.Info("room-service (gRPC) listening", "port", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve grpc", "error", err)
	}
}
