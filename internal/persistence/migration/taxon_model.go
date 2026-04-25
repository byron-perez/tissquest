package migration

import "gorm.io/gorm"

type TaxonModel struct {
	gorm.Model
	Rank      string         `gorm:"index"`
	Name      string
	ParentID  *uint          `gorm:"index"`
	Parent    *TaxonModel    `gorm:"foreignKey:ParentID"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (TaxonModel) TableName() string {
	return "taxa"
}
