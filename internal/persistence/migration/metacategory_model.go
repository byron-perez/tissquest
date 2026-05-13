package migration

import "gorm.io/gorm"

type MetacategoryModel struct {
	gorm.Model
	Name        string
	Description string
	ParentID    *uint          `gorm:"index"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Parent      *MetacategoryModel
	Children    []MetacategoryModel `gorm:"foreignKey:ParentID"`
	Categories  []CategoryModel     `gorm:"foreignKey:MetacategoryID"`
}

func (MetacategoryModel) TableName() string {
	return "metacategories"
}
