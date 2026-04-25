package migration

import "gorm.io/gorm"

type SlideModel struct {
	gorm.Model
	Name           string
	ImageKey       string         // logical image identity, e.g. "slides/6"
	TissueRecordID uint           `gorm:"index"`
	PreparationID  uint           `gorm:"index"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	Preparation    PreparationModel
	Magnification  int

	// Virtual microscope fields — all optional (zero/empty = not tiled yet).
	DziURL            string  `gorm:"column:dzi_url"`
	BaseMagnification int     `gorm:"column:base_magnification"`
	MicronsPerPixel   float64 `gorm:"column:microns_per_pixel"`
	HomeViewport      string  `gorm:"column:home_viewport;type:text"` // JSON-encoded ViewportPosition, empty = unset
}

func (SlideModel) TableName() string {
	return "slides"
}
