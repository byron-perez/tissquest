package services

import (
	"fmt"
	"mime/multipart"

	corestorage "mcba/tissquest/internal/core/storage"
)

type SlideService struct {
	storage corestorage.ImageStorage
}

func NewSlideService(storage corestorage.ImageStorage) *SlideService {
	return &SlideService{storage: storage}
}

func (s *SlideService) UploadImage(slideID uint, file multipart.File, header *multipart.FileHeader) (string, error) {
	data := make([]byte, header.Size)
	if _, err := file.Read(data); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	filename := fmt.Sprintf("slides/%d-%s", slideID, header.Filename)
	return s.storage.Upload(filename, contentType, data)
}
