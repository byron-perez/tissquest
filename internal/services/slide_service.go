package services

import (
	"fmt"
	"mime/multipart"

	"mcba/tissquest/internal/core/slide"
	corestorage "mcba/tissquest/internal/core/storage"
)

type SlideService struct {
	storage   corestorage.ImageStorage
	slideRepo slide.RepositoryInterface
}

func NewSlideService(storage corestorage.ImageStorage, repo slide.RepositoryInterface) *SlideService {
	return &SlideService{storage: storage, slideRepo: repo}
}

func (s *SlideService) UploadImage(slideID uint, file multipart.File, header *multipart.FileHeader) (string, error) {
	data := make([]byte, header.Size)
	if _, err := file.Read(data); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}

	// Store under slides/original/ so Lambda trigger can process it
	filename := fmt.Sprintf("slides/original/%d.png", slideID)
	url, err := s.storage.Upload(filename, contentType, data)
	if err != nil {
		return "", err
	}

	// Record the original variant and set the logical image key
	imageKey := fmt.Sprintf("slides/%d", slideID)
	if err := s.slideRepo.SetImageVariant(slideID, slide.ImageSizeOriginal, url); err != nil {
		return "", fmt.Errorf("failed to store image variant: %w", err)
	}

	// Update the slide's ImageKey so it knows it has an image
	sl, err := s.slideRepo.GetByID(slideID)
	if err != nil {
		return "", err
	}
	sl.ImageKey = imageKey
	if err := s.slideRepo.Update(slideID, sl); err != nil {
		return "", fmt.Errorf("failed to update image key: %w", err)
	}

	return url, nil
}

func (s *SlideService) SetImageVariant(slideID uint, size slide.ImageSize, url string) error {
	return s.slideRepo.SetImageVariant(slideID, size, url)
}

func (s *SlideService) Create(tissueRecordID uint, sl *slide.Slide) (uint, error) {
	if err := sl.Validate(); err != nil {
		return 0, err
	}
	sl.TissueRecordID = tissueRecordID
	return s.slideRepo.Save(sl)
}

func (s *SlideService) GetByID(id uint) (*slide.Slide, error) {
	return s.slideRepo.GetByID(id)
}

func (s *SlideService) Update(id uint, sl *slide.Slide) error {
	if err := sl.Validate(); err != nil {
		return err
	}
	return s.slideRepo.Update(id, sl)
}

func (s *SlideService) Delete(id uint) error {
	return s.slideRepo.Delete(id)
}

func (s *SlideService) ListByTissueRecord(tissueRecordID uint) ([]slide.Slide, error) {
	return s.slideRepo.ListByTissueRecord(tissueRecordID)
}

func (s *SlideService) ListDisplayByTissueRecord(tissueRecordID uint, preferredSize slide.ImageSize) ([]slide.DisplaySlide, error) {
	return s.slideRepo.ListDisplayByTissueRecord(tissueRecordID, preferredSize)
}

// SetDziMetadata is called by the tiling pipeline CLI after a slide has been processed.
func (s *SlideService) SetDziMetadata(slideID uint, dziURL string, baseMagnification int, micronsPerPixel float64) error {
	return s.slideRepo.SetDziMetadata(slideID, dziURL, baseMagnification, micronsPerPixel)
}
