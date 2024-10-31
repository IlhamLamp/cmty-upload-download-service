package utils

import (
	"context"
	"log"
	"mime/multipart"
	"path/filepath"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

type CloudinaryClient struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryClient(cloudName, apiKey, apiSecret string) (*CloudinaryClient, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &CloudinaryClient{cld: cld}, nil
}

func (c *CloudinaryClient) UploadFile(file multipart.File, originalFileName string) (string, error) {
	ctx := context.Background()
	uniqueFileName := uuid.New().String() + filepath.Ext(originalFileName)
	uploadParams := uploader.UploadParams{
		PublicID: uniqueFileName,
		Folder:   "uploads",
	}
	uploadResult, err := c.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "An error occured wher upload", err
	}
	return uploadResult.SecureURL, nil
}

func (c *CloudinaryClient) DeleteFile(publicId string) error {
	_, err := c.cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: publicId})
	if err != nil {
		log.Printf("Failed to delete image from cloudinary: %v", err)
		return err
	}
	log.Printf("Successfully deleted image with Public ID: %s", publicId)
	return nil
}
