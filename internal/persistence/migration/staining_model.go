package migration

import "gorm.io/gorm"

type StainingModel struct {
	gorm.Model
	ID   uint `gorm:"primaryKey"`
	Name string
}

func (StainingModel) TableName() string {
	return "stainings"
}
