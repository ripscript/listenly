package config

type MusicConfig struct {
	GRPCPort         string
	DatabaseURL      string
	AuthServiceAddr  string
	RoomServiceAddr  string
	MediaServiceAddr string
	ElasticsearchURL string
	RabbitMQURL      string
}

func LoadMusicConfig() MusicConfig {
	return MusicConfig{
		GRPCPort:         getEnv("GRPC_PORT", "50053"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:54322/listenly_music?sslmode=disable"),
		AuthServiceAddr:  getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		RoomServiceAddr:  getEnv("ROOM_SERVICE_ADDR", "localhost:50052"),
		MediaServiceAddr: getEnv("MEDIA_SERVICE_ADDR", "localhost:50054"),
		ElasticsearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://admin:admin123@localhost:5672/"),
	}
}
