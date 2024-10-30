package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	CloudinaryApiKey    string
	CloudinaryApiSecret string
	CloudinaryCloudName string
	JwtAccessSecret     string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading env file")
	}

	config := Config{
		CloudinaryApiKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryApiSecret: os.Getenv("CLOUDINARY_API_SECRET"),
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		JwtAccessSecret:     os.Getenv("JWT_ACCESS_SECRET"),
	}

	if config.CloudinaryApiKey == "" || config.CloudinaryApiSecret == "" ||
		config.CloudinaryCloudName == "" || config.JwtAccessSecret == "" {
		log.Fatal("Required environment variable missing")
	}

	return config
}
