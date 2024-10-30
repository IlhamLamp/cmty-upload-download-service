package handlers

import (
	"go-upload-download-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadResponse struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	SecureURL string `json:"secure_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

func UploadFileHandler(cldClient *utils.CloudinaryClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, UploadResponse{
				Status:  http.StatusBadRequest,
				Message: "File is required",
				Error:   "File is required",
			})
			return
		}

		openedFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, UploadResponse{
				Status:  http.StatusInternalServerError,
				Message: "Cannot open file",
				Error:   "Cannot open file",
			})
			return
		}
		defer openedFile.Close()

		fileName := file.Filename
		url, err := cldClient.UploadFile(openedFile, fileName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, UploadResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to upload to Cloudinary",
				Error:   "Failed to upload to Cloudinary",
			})
			return
		}

		c.JSON(http.StatusOK, UploadResponse{
			Status:    http.StatusOK,
			Message:   "File uploaded successfully",
			SecureURL: url,
		})
	}
}
