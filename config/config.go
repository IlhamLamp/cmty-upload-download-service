package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort             string
	CloudinaryApiKey    string
	CloudinaryApiSecret string
	CloudinaryCloudName string
	JwtAccessSecret     string
	RabbitMQUrl         string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading env file")
	}

	requiredEnv := []string{
		"APP_PORT",
		"CLOUDINARY_API_KEY",
		"CLOUDINARY_API_SECRET",
		"CLOUDINARY_CLOUD_NAME",
		"JWT_ACCESS_SECRET",
		"RABBITMQ_URL",
	}

	missingEnv := checkRequiredEnv(requiredEnv)
	if len(missingEnv) > 0 {
		log.Fatalf("Missing required environment variables: %v", missingEnv)
	}

	config := Config{
		AppPort:             os.Getenv("APP_PORT"),
		CloudinaryApiKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryApiSecret: os.Getenv("CLOUDINARY_API_SECRET"),
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		JwtAccessSecret:     os.Getenv("JWT_ACCESS_SECRET"),
		RabbitMQUrl:         os.Getenv("RABBITMQ_URL"),
	}

	return config
}

func checkRequiredEnv(keys []string) []string {
	var missing []string
	for _, key := range keys {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}
