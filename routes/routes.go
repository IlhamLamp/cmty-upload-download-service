package routes

import (
	"go-upload-download-service/handlers"
	"go-upload-download-service/middleware"
	"go-upload-download-service/utils"

	"github.com/gin-gonic/gin"
)

func RegisterUploadRoutes(router *gin.RouterGroup, cldClient *utils.CloudinaryClient, mqClient *utils.RabbitMQClient ,jwtSecret string) {
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Upload-download service")
	})
	router.POST("/upload", middleware.AuthMiddleware(jwtSecret), handlers.UploadFileHandler(cldClient, mqClient))
}
