package migration

import "gorm.io/gorm"

type TaxonModel struct {
	gorm.Model
	Rank     string
	Name     string
	ParentID *uint
	Parent   *TaxonModel `gorm:"foreignKey:ParentID"`
}

func (TaxonModel) TableName() string {
	return "taxa"
}
