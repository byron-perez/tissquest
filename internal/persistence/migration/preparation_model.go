package migration

import "gorm.io/gorm"

type PreparationModel struct {
	gorm.Model
	Staining        string
	InclusionMethod string
	Reagents        string
	Protocol        string
	Notes           string
}

func (PreparationModel) TableName() string {
	return "preparations"
}
