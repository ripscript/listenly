package config

type RoomConfig struct {
	GRPCPort        string
	DatabaseURL     string
	AuthServiceAddr string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	RabbitMQURL string
}

func LoadRoomConfig() RoomConfig {
	return RoomConfig{
		GRPCPort:        getEnv("GRPC_PORT", "50052"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:54322/listenly_room?sslmode=disable"),
		AuthServiceAddr: getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,

		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://admin:admin123@localhost:5672/"),
	}
}
