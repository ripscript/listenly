package config

type GatewayConfig struct {
	HTTPPort         string
	AuthServiceAddr  string
	RoomServiceAddr  string
	MusicServiceAddr string
	WebBaseURL       string
	RabbitMQURL      string
}

func LoadGatewayConfig() GatewayConfig {
	return GatewayConfig{
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		AuthServiceAddr:  getEnv("AUTH_SERVICE_ADDR", "auth-service:50051"),
		RoomServiceAddr:  getEnv("ROOM_SERVICE_ADDR", "room-service:50052"),
		MusicServiceAddr: getEnv("MUSIC_SERVICE_ADDR", "localhost:50053"),
		WebBaseURL:       getEnv("WEB_BASE_URL", "http://localhost:3000"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://admin:admin123@localhost:5672/"),
	}
}
