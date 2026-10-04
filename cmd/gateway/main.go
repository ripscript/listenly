package main

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"

	"listenly-backend/config"
	"listenly-backend/internal/gateway/handler"
	"listenly-backend/internal/gateway/middleware"
	"listenly-backend/internal/gateway/ws"
	"listenly-backend/pkg/grpcclient"
	"listenly-backend/pkg/rabbitmq"
	"listenly-backend/pkg/validator"
)

func main() {
	config.LoadEnv("cmd/gateway/.env")
	cfg := config.LoadGatewayConfig()

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
	musicClient, err := grpcclient.NewMusicServiceClient(cfg.MusicServiceAddr)
	if err != nil {
		slog.Error("failed to connect music-service", "error", err)
		return
	}

	// --- WebSocket realtime ---
	hub := ws.NewHub()

	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQURL, []string{"room.*.playback", "room.*.queue"})
	if err != nil {
		slog.Error("failed to connect rabbitmq consumer", "error", err)
		return
	}
	defer consumer.Close()

	if err := ws.StartDispatcher(hub, consumer); err != nil {
		slog.Error("failed to start ws dispatcher", "error", err)
		return
	}

	wsHandler := ws.NewHandler(hub, authClient)

	authHandler := handler.NewAuthHandler(authClient)
	roomHandler := handler.NewRoomHandler(roomClient, cfg.WebBaseURL)
	musicHandler := handler.NewMusicHandler(musicClient)

	e := echo.NewWithConfig(echo.Config{
		NoGroupAutoRegister404Routes: true,
	})
	e.Validator = validator.New()
	e.Use(echomw.RequestLogger())
	e.Use(echomw.Recover())

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	api := e.Group("/api")

	v1 := api.Group("/v1")

	v1.GET("/rooms/:uuid/ws", wsHandler.Connect)

	authGroup := v1.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/logout", authHandler.Logout)

	protected := v1.Group("", middleware.JWTAuth(authClient))
	protected.GET("/auth/me", authHandler.Verify)
	protected.POST("/auth/logout-all", authHandler.LogoutAll)

	roomGroup := protected.Group("/rooms")
	roomGroup.POST("", roomHandler.CreateRoom)
	roomGroup.GET("/:uuid", roomHandler.GetRoom)
	roomGroup.POST("/join", roomHandler.JoinPrivateRoom)
	roomGroup.DELETE("/:uuid/leave", roomHandler.LeaveRoom)
	roomGroup.GET("", roomHandler.ListPublicRooms)
	roomGroup.POST("/:uuid/join", roomHandler.JoinRoom)
	roomGroup.GET("/me", roomHandler.ListMyRooms)
	roomGroup.PUT("/:uuid/playback", roomHandler.UpdatePlayback)
	roomGroup.GET("/:uuid/playback", roomHandler.GetPlayback)

	musicGroup := protected.Group("/music")
	musicGroup.GET("/search", musicHandler.SearchTracks)
	musicGroup.POST("/rooms/:roomUuid/queue", musicHandler.RequestTrack)
	musicGroup.GET("/rooms/:roomUuid/queue", musicHandler.GetQueue)
	musicGroup.DELETE("/rooms/:roomUuid/queue/:uuid", musicHandler.RemoveFromQueue)
	musicGroup.PATCH("/queue/:uuid/played", musicHandler.MarkAsPlayed)
	musicGroup.GET("/tracks/:uuid", musicHandler.GetTrack)
	musicGroup.GET("/tracks/:uuid/stream", musicHandler.GetStreamURL)
	musicGroup.GET("/youtube/search", musicHandler.SearchYouTube)
	musicGroup.POST("/rooms/:roomUuid/queue/advance", musicHandler.AdvanceQueue)

	if err := e.Start(":" + cfg.HTTPPort); err != nil {
		slog.Error("failed to start gateway", "error", err)
	}
}
