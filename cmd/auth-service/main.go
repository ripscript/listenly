package main

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"listenly-backend/config"
	authv1 "listenly-backend/gen/go/auth/v1"
	"listenly-backend/internal/auth/handler"
	"listenly-backend/internal/auth/repository"
	"listenly-backend/internal/auth/service"
	"listenly-backend/pkg/database"
	"listenly-backend/pkg/grpcserver"
	"listenly-backend/pkg/jwt"
)

func main() {
	config.LoadEnv("cmd/auth-service/.env")
	cfg := config.LoadAuthConfig()

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		return
	}

	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	authRepo := repository.NewAuthRepository(db)
	authSvc := service.NewAuthService(authRepo, jwtManager)
	authHandler := handler.NewAuthGRPCHandler(authSvc)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.RecoveryInterceptor),
	)
	authv1.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	slog.Info("auth-service (gRPC) listening", "port", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve grpc", "error", err)
	}
}
