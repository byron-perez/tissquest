package migration

import "gorm.io/gorm"

type CategoryModel struct {
    gorm.Model
    ID            uint           `gorm:"primaryKey"`
    Name          string
    Type          string
    Description   string
    ParentID      *uint
    Parent        *CategoryModel
    Children      []CategoryModel `gorm:"foreignKey:ParentID"`
    TissueRecords []TissueRecordModel `gorm:"many2many:tissue_record_categories;"`
    Atlases       []AtlasModel        `gorm:"many2many:atlas_categories;"`
}

func (CategoryModel) TableName() string {
    return "categories"
}
