package services

import (
	"mcba/tissquest/internal/core/slide"
)

type SlideService struct {
	slideRepo slide.RepositoryInterface
}

func NewSlideService(repo slide.RepositoryInterface) *SlideService {
	return &SlideService{slideRepo: repo}
}

func (s *SlideService) SetImageVariant(slideID uint, size slide.ImageSize, url string) error {
	return s.slideRepo.SetImageVariant(slideID, size, url)
}

func (s *SlideService) SetImageKey(slideID uint, imageKey string) error {
	return s.slideRepo.SetImageKey(slideID, imageKey)
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

// GetRandomTiledDisplaySlides returns up to limit slides that have been tiled, in random order.
func (s *SlideService) GetRandomTiledDisplaySlides(limit int) ([]slide.DisplaySlide, error) {
	return s.slideRepo.GetRandomTiledDisplaySlides(limit)
}

// SetDziMetadata is called by the tiling pipeline CLI after a slide has been processed.
func (s *SlideService) SetDziMetadata(slideID uint, dziURL string, baseMagnification int, micronsPerPixel float64) error {
	return s.slideRepo.SetDziMetadata(slideID, dziURL, baseMagnification, micronsPerPixel)
}

// SetSourceMetadata stores provenance fields for a slide (source URL, title, author, license, notes).
func (s *SlideService) SetSourceMetadata(slideID uint, sourceURL string, sourceTitle string, sourceAuthor string, sourceLicense string, sourceNotes string) error {
	return s.slideRepo.SetSourceMetadata(slideID, sourceURL, sourceTitle, sourceAuthor, sourceLicense, sourceNotes)
}
