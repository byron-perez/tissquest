package migration

import "gorm.io/gorm"

type SlideModel struct {
	gorm.Model
	Name           string
	Url            string
	TissueRecordID uint
	PreparationID  uint
	Preparation    PreparationModel
	Magnification  int
}

func (SlideModel) TableName() string {
	return "slides"
}
