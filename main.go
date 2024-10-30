package main

import (
	"go-upload-download-service/config"
	"go-upload-download-service/middleware"
	"go-upload-download-service/routes"
	"go-upload-download-service/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.LoadConfig()
	cldClient, err := utils.NewCloudinaryClient(conf.CloudinaryCloudName, conf.CloudinaryApiKey, conf.CloudinaryApiSecret)
	if err != nil {
		log.Fatalf("Failed to create Cloudinary client: %v", err)
	}
	router := gin.New()
	router.Use(middleware.CORSMiddleware())

	api := router.Group("/api/v1")
	routes.RegisterUploadRoutes(api, cldClient, conf.JwtAccessSecret)

	router.Run(":3100")
}
