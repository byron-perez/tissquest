package migration

import "gorm.io/gorm"

type CategoryModel struct {
    gorm.Model
    Name          string
    Type          string
    Description   string
    ParentID      *uint
    Parent        *CategoryModel
    Children      []CategoryModel `gorm:"foreignKey:ParentID"`
    TissueRecords []TissueRecordModel `gorm:"many2many:tissue_record_categories;joinForeignKey:CategoryID;joinReferences:TissueRecordID"`
}

func (CategoryModel) TableName() string {
    return "categories"
}
