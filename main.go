package main

import (
	"github.com/gin-gonic/gin"
	"go-upload-download-service/config"
	"go-upload-download-service/middleware"
	"go-upload-download-service/routes"
	"go-upload-download-service/utils"
	"go-upload-download-service/workers"
	"log"
)

func main() {
	conf := config.LoadConfig()
	rabbitMqUrl := conf.GetRabbitMQUrl()

	cldClient, err := utils.NewCloudinaryClient(conf.CloudinaryCloudName, conf.CloudinaryApiKey, conf.CloudinaryApiSecret)
	if err != nil {
		log.Fatalf("Failed to create Cloudinary client: %v", err)
	}

	mqClient, err := utils.NewRabbitMQClient(rabbitMqUrl, "delete_image_cloudinary")
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ client: %v", err)
	}
	defer mqClient.Close()

	go workers.StartDeleteImageWorker(mqClient, cldClient)

	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Logger())

	uploadDeps := routes.UploadServiceDeps{
		CloudinaryClient: cldClient,
		RabbitMqClient:   mqClient,
		JwtSecret:        conf.JwtAccessSecret,
		AppName:          conf.AppName,
	}

	api := router.Group("/api/v1/media")
	routes.RegisterUploadRoutes(api, uploadDeps)

	router.Run(":" + conf.AppPort)
}
