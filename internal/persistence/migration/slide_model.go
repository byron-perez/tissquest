package migration

import "gorm.io/gorm"

type SlideModel struct {
	gorm.Model
	ID             uint `gorm:"primaryKey"`
	Name           string
	TissueRecordID uint
}

func (SlideModel) TableName() string {
	return "slides"
}
