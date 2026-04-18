package slide

import "errors"

var (
	ErrEmptyName            = errors.New("slide name cannot be empty")
	ErrInvalidMagnification = errors.New("magnification must be a positive value")
)

type Preparation struct {
	Staining        string
	InclusionMethod string
	Reagents        string
	Protocol        string
	Notes           string
}

// Slide is the domain object. It knows nothing about image storage or resolutions.
// ImageKey is the logical identifier used by the storage layer to locate variants
// (e.g. "slides/6"). Empty means no image has been uploaded yet.
type Slide struct {
	ID             uint
	TissueRecordID uint
	Name           string
	ImageKey       string // logical image identity, e.g. "slides/6"
	Magnification  int
	Preparation    Preparation
}

func (s *Slide) Validate() error {
	if s.Name == "" {
		return ErrEmptyName
	}
	if s.Magnification <= 0 {
		return ErrInvalidMagnification
	}
	return nil
}
