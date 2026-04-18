package migration

import "gorm.io/gorm"

// SlideImageVariantModel stores one URL per (slide, size) pair.
// New sizes can be added as rows without any schema change.
type SlideImageVariantModel struct {
	gorm.Model
	SlideID uint   `gorm:"not null;index"`
	Size    string `gorm:"not null"` // "original" | "low" | "medium"
	Url     string `gorm:"not null"`
}

func (SlideImageVariantModel) TableName() string {
	return "slide_image_variants"
}
