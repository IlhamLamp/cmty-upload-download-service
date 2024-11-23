package routes

import (
	"go-upload-download-service/handlers"
	"go-upload-download-service/middleware"
	"go-upload-download-service/utils"

	"github.com/gin-gonic/gin"
)

type UploadServiceDeps struct {
	CloudinaryClient *utils.CloudinaryClient
	RabbitMqClient   *utils.RabbitMQClient
	JwtSecret        string
	AppName          string
}

func RegisterUploadRoutes(router *gin.RouterGroup, deps UploadServiceDeps) {
	router.GET("/", func(c *gin.Context) {
		c.String(200, deps.AppName)
	})
	router.POST("/upload", middleware.AuthMiddleware(deps.JwtSecret), handlers.UploadFileHandler(deps.CloudinaryClient, deps.RabbitMqClient))
}
