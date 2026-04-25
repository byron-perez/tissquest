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

// ViewportPosition holds a saved OpenSeadragon viewport state.
// Coordinates are normalized (0.0–1.0 relative to image width).
// Zoom is a multiplier relative to the fit-to-screen state.
type ViewportPosition struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

// Slide is the domain object. It knows nothing about image storage or resolutions.
// ImageKey is the logical identifier used by the storage layer to locate variants
// (e.g. "slides/6"). Empty means no image has been uploaded yet.
//
// DziURL, BaseMagnification, MicronsPerPixel and HomeViewport are optional.
// A zero/nil value means the slide has not been processed for tiled viewing yet;
// the UI falls back to the static image in that case.
type Slide struct {
	ID             uint
	TissueRecordID uint
	Name           string
	ImageKey       string // logical image identity, e.g. "slides/6"
	Magnification  int

	// Virtual microscope fields — all optional.
	DziURL            string            // S3 URL to the .dzi descriptor; empty = not tiled yet
	BaseMagnification int               // objective used at capture (4, 10, 40, 100)
	MicronsPerPixel   float64           // spatial calibration: µm per image pixel
	HomeViewport      *ViewportPosition // curated starting position; nil = fit-to-screen

	Preparation Preparation
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

// IsTiled reports whether this slide has been processed into a tiled image.
func (s *Slide) IsTiled() bool {
	return s.DziURL != ""
}
