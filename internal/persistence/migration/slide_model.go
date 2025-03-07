package migration

import "gorm.io/gorm"

type SlideModel struct {
	gorm.Model
	ID             uint `gorm:"primaryKey"`
	Name           string
	Url            string
	TissueRecordID uint
	StainingID     uint
	Staining       StainingModel
}

func (SlideModel) TableName() string {
	return "slides"
}
