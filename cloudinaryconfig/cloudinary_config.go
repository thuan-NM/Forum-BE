package config

import (
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"os"
)

// CloudinaryConfig chứa cấu hình Cloudinary
type CloudinaryConfig struct {
	CloudName    string
	APIKey       string
	APISecret    string
	UploadPreset string
}

// NewCloudinaryClient khởi tạo client Cloudinary
func NewCloudinaryClient() (*cloudinary.Cloudinary, error) {
	config := CloudinaryConfig{
		CloudName:    os.Getenv("CLOUDINARY_CLOUD_NAME"),
		APIKey:       os.Getenv("CLOUDINARY_API_KEY"),
		APISecret:    os.Getenv("CLOUDINARY_API_SECRET"),
		UploadPreset: os.Getenv("CLOUDINARY_UPLOAD_PRESET"),
	}

	if config.CloudName == "" || config.APIKey == "" || config.APISecret == "" {
		return nil, errors.New("missing Cloudinary configuration")
	}

	cld, err := cloudinary.NewFromParams(config.CloudName, config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	return cld, nil
}
