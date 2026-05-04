package slide

// ImageSize represents a named resolution variant of a slide image.
type ImageSize string

const (
	ImageSizeOriginal ImageSize = "original"
	ImageSizeThumb    ImageSize = "thumb"
	ImageSizePreview  ImageSize = "preview"
)

// DisplaySlide is a read model for the presentation layer.
// It carries the most appropriate image URL for the requested context,
// assembled by the repository — the rest of the app never picks a size.
type DisplaySlide struct {
	Slide
	ImageUrl string // resolved URL for the requested size (falls back gracefully)
}

type RepositoryInterface interface {
	Save(sl *Slide) (uint, error)
	GetByID(id uint) (*Slide, error)
	Update(id uint, sl *Slide) error
	// SetImageVariant stores a URL for a specific size variant of a slide's image.
	// Called by the Lambda callback endpoint after processing.
	SetImageVariant(slideID uint, size ImageSize, url string) error
	// SetDziMetadata persists the tiled-image fields after the pipeline runs.
	SetDziMetadata(slideID uint, dziURL string, baseMagnification int, micronsPerPixel float64) error
	// GetPendingTiling returns all slides that have a source image but no DZI yet.
	GetPendingTiling() ([]Slide, error)
	// GetRandomTiledDisplaySlides returns up to limit slides that have a DZI, in random order.
	GetRandomTiledDisplaySlides(limit int) ([]DisplaySlide, error)
	Delete(id uint) error
	ListByTissueRecord(tissueRecordID uint) ([]Slide, error)
	// ListDisplayByTissueRecord returns slides with the best available image URL
	// for the given preferred size (falls back to original if preferred not available).
	ListDisplayByTissueRecord(tissueRecordID uint, preferredSize ImageSize) ([]DisplaySlide, error)
}
