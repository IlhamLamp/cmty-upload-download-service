package utils

import (
	"context"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
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

func (c *CloudinaryClient) UploadFile(file multipart.File, fileName string) (string, error) {
	ctx := context.Background()
	uploadParams := uploader.UploadParams{
		PublicID: fileName,
		Folder:   "uploads",
	}
	uploadResult, err := c.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "An error occured wher upload", err
	}
	return uploadResult.SecureURL, nil
}
