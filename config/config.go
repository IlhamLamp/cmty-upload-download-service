package config

import (
	_ "github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	AppClient           string
	AppPort             string
	AppName             string
	CloudinaryApiKey    string
	CloudinaryApiSecret string
	CloudinaryCloudName string
	JwtAccessSecret     string
	JwtRefreshSecret    string
	RabbitMQUser        string
	RabbitMQPassword    string
	RabbitMQHost        string
	RabbitMQPort        string
}

func LoadConfig() Config {
	// UNCOMMENT FOR RUNNING LOCALLY
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading env file")
	// }

	requiredEnv := []string{
		"APP_CLIENT",
		"APP_PORT",
		"APP_NAME",
		"CLOUDINARY_API_KEY",
		"CLOUDINARY_API_SECRET",
		"CLOUDINARY_CLOUD_NAME",
		"JWT_ACCESS_SECRET",
		"JWT_REFRESH_SECRET",
		"RABBITMQ_USER",
		"RABBITMQ_PASSWORD",
		"RABBITMQ_HOST",
		"RABBITMQ_PORT",
	}

	missingEnv := checkRequiredEnv(requiredEnv)
	if len(missingEnv) > 0 {
		log.Fatalf("Missing required environment variables: %v", missingEnv)
	}

	config := Config{
		AppClient:           os.Getenv("APP_CLIENT"),
		AppPort:             os.Getenv("APP_PORT"),
		AppName:             os.Getenv("APP_NAME"),
		CloudinaryApiKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryApiSecret: os.Getenv("CLOUDINARY_API_SECRET"),
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		JwtAccessSecret:     os.Getenv("JWT_ACCESS_SECRET"),
		JwtRefreshSecret:    os.Getenv("JWT_REFRESH_SECRET"),
		RabbitMQUser:        os.Getenv("RABBITMQ_USER"),
		RabbitMQPassword:    os.Getenv("RABBITMQ_PASSWORD"),
		RabbitMQHost:        os.Getenv("RABBITMQ_HOST"),
		RabbitMQPort:        os.Getenv("RABBITMQ_PORT"),
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

func (c *Config) GetRabbitMQUrl() string {
	if c.RabbitMQHost == "" {
		c.RabbitMQHost = "localhost"
	}
	if c.RabbitMQPort == "" {
		c.RabbitMQPort = "5672"
	}
	url := "amqp://" + c.RabbitMQUser + ":" + c.RabbitMQPassword + "@" + c.RabbitMQHost + ":" + c.RabbitMQPort + "/"
	return url
}
