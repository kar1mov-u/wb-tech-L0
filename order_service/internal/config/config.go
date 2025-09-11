package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	MaxConns int32 `yaml:"max_conn" env:"MAX_CONN" env-default:"10"`
	MinConns int32 `yaml:"min_conn" env:"MIN_CONN" env-default:"5"`
}

type RedisConfig struct {
	ConnString string
}

type KafkaConfig struct {
	Host    string
	Port    int
	GroupID string
	Topic   string
	Broker  string
}

func LoadConfig() Config {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found")
	}
	// Parse configuration
	return Config{
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "user"),
			Password: getEnv("POSTGRES_PASSWORD", "password"),
			Database: getEnv("POSTGRES_DB", "orders"),
		},
		Redis: RedisConfig{
			ConnString: getEnv("REDIS_CONN_STRING", "redis://localhost:6379/0"),
		},
		Kafka: KafkaConfig{
			Host:    getEnv("RABBITMQ_HOST", "localhost"),
			Port:    getEnvAsInt("RABBITMQ_PORT", 5672),
			GroupID: getEnv("KAFKA_GROUP_ID", "group1"),
			Topic:   getEnv("KAFKA_TOPIC", "orders"),
			Broker:  getEnv("KAFKA_BROKER", "localhost:9092"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
