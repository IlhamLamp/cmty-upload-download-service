package routes

import (
	"github.com/gin-gonic/gin"
	"go-upload-download-service/handlers"
	"go-upload-download-service/middleware"
	"go-upload-download-service/utils"
)

type UploadServiceDeps struct {
	CloudinaryClient *utils.CloudinaryClient
	RabbitMqClient   *utils.RabbitMQClient
	JwtSecret        string
	AppName          string
}

func RegisterUploadRoutes(router *gin.RouterGroup, deps UploadServiceDeps) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service_name": deps.AppName,
			"description":  "Media upload download service",
		})
	})
	router.POST("/upload", middleware.AuthMiddleware(deps.JwtSecret), handlers.UploadFileHandler(deps.CloudinaryClient, deps.RabbitMqClient))
}
