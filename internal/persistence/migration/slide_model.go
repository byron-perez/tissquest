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
}

func (SlideModel) TableName() string {
	return "slides"
}
