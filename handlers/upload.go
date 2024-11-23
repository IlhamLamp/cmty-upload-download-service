package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-upload-download-service/utils"
	"log"
	"net/http"
	"regexp"
)

type UploadResponse struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	SecureURL string `json:"secure_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

func respondWithError(c *gin.Context, status int, message string) {
	c.JSON(status, UploadResponse{
		Status:  status,
		Message: message,
		Error:   message,
	})
}

func extractPublicID(oldImageURL string) (string, error) {
	re := regexp.MustCompile(`/([^/]+/[^/.]+\.[a-z]+)`)
	match := re.FindStringSubmatch(oldImageURL)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", fmt.Errorf("failed to extract public ID from URL: %s", oldImageURL)
}

func UploadFileHandler(cldClient *utils.CloudinaryClient, mqClient *utils.RabbitMQClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Start handling file upload request")
		file, err := c.FormFile("file")
		if err != nil {
			log.Printf("Error fetching file from request: %v", err)
			respondWithError(c, http.StatusBadRequest, "File is required!")
			return
		}
		log.Printf("File received: %s", file.Filename)

		openedFile, err := file.Open()
		if err != nil {
			log.Printf("Error opening file: %v", err)
			respondWithError(c, http.StatusInternalServerError, "Cannot open file!")
			return
		}
		defer openedFile.Close()

		fileName := file.Filename

		log.Println("Uploading file to Cloudinary...")
		url, err := cldClient.UploadFile(openedFile, fileName)
		if err != nil {
			log.Printf("Error uploading to Cloudinary: %v", err)
			respondWithError(c, http.StatusInternalServerError, "Failed to upload into cloudinary")
			return
		}
		log.Printf("File uploaded successfully to: %s", url)

		oldImageURL := c.PostForm("old_image_url")

		if oldImageURL != "" {
			if publicId, err := extractPublicID(oldImageURL); err == nil {
				if err := mqClient.PublishDeleteImageMessage(publicId); err != nil {
					respondWithError(c, http.StatusInternalServerError, "Failed to queue old image for deletion")
					return
				}
			} else {
				log.Printf("Error extracting public ID: %v", err)
			}
		}

		c.JSON(http.StatusOK, UploadResponse{
			Status:    http.StatusOK,
			Message:   "File uploaded successfully",
			SecureURL: url,
		})
	}
}
