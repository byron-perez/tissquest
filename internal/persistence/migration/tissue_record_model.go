package migration

import "gorm.io/gorm"

type TissueRecordModel struct {
	gorm.Model
	ID             uint `gorm:"primaryKey"`
	Name           string
	Notes          string
	Taxonomicclass string
	Slides         []SlideModel `gorm:"foreignKey:TissueRecordID;"`
}

func (TissueRecordModel) TableName() string {
	return "tissue_records"
}
