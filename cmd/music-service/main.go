package main

import (
	"context"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"listenly-backend/config"
	musicv1 "listenly-backend/gen/go/music/v1"
	"listenly-backend/internal/music/handler"
	"listenly-backend/internal/music/repository"
	"listenly-backend/internal/music/service"
	"listenly-backend/pkg/database"
	"listenly-backend/pkg/elastic"
	"listenly-backend/pkg/grpcclient"
	"listenly-backend/pkg/grpcserver"
)

func main() {
	config.LoadEnv("cmd/music-service/.env")
	cfg := config.LoadMusicConfig()

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		return
	}

	esClient, err := elastic.NewClient([]string{cfg.ElasticsearchURL})
	if err != nil {
		slog.Error("failed to connect elasticsearch", "error", err)
		return
	}

	authClient, err := grpcclient.NewAuthServiceClient(cfg.AuthServiceAddr)
	if err != nil {
		slog.Error("failed to connect auth-service", "error", err)
		return
	}

	roomClient, err := grpcclient.NewRoomServiceClient(cfg.RoomServiceAddr)
	if err != nil {
		slog.Error("failed to connect room-service", "error", err)
		return
	}

	mediaClient, err := grpcclient.NewMediaServiceClient(cfg.MediaServiceAddr)
	if err != nil {
		slog.Error("failed to connect media-service", "error", err)
		return
	}

	musicRepo := repository.NewMusicRepository(db)
	trackSearchRepo := repository.NewTrackSearchRepository(esClient)

	// PASTIKAN BARIS INI ADA — inilah yang tadi hilang
	if count, err := service.ReindexAllTracks(context.Background(), musicRepo, trackSearchRepo); err != nil {
		slog.Warn("failed to reindex tracks on startup", "error", err)
	} else {
		slog.Info("reindexed tracks on startup", "count", count)
	}

	musicSvc := service.NewMusicService(musicRepo, trackSearchRepo, authClient, roomClient, mediaClient)
	musicHandler := handler.NewMusicGRPCHandler(musicSvc)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.RecoveryInterceptor),
	)
	musicv1.RegisterMusicServiceServer(grpcServer, musicHandler)
	reflection.Register(grpcServer)

	slog.Info("music-service (gRPC) listening", "port", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve grpc", "error", err)
	}
}
