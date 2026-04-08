package migration

import "gorm.io/gorm"

type TissueRecordModel struct {
    gorm.Model
    ID             uint `gorm:"primaryKey"`
    Name           string
    Notes          string
    Taxonomicclass string
    Slides         []SlideModel   `gorm:"foreignKey:TissueRecordID;"`
    Categories     []CategoryModel `gorm:"many2many:tissue_record_categories;"`
    Atlases        []AtlasModel   `gorm:"many2many:atlas_tissue_records;"`
}

func (TissueRecordModel) TableName() string {
    return "tissue_records"
}
