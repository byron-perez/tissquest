package migration

import "gorm.io/gorm"

type CategoryModel struct {
	gorm.Model
	Name           string
	Type           string `gorm:"index"`
	Description    string
	MetacategoryID *uint          `gorm:"index"`
	ParentID       *uint          `gorm:"index"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	Metacategory   *MetacategoryModel
	Parent         *CategoryModel
	Children       []CategoryModel     `gorm:"foreignKey:ParentID"`
	TissueRecords  []TissueRecordModel `gorm:"many2many:tissue_record_categories;joinForeignKey:CategoryID;joinReferences:TissueRecordID"`
}

func (CategoryModel) TableName() string {
	return "categories"
}
