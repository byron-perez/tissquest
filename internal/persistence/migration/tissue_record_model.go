package migration

import "gorm.io/gorm"

type TissueRecordModel struct {
    gorm.Model
    ID             uint `gorm:"primaryKey"`
    Name           string
    Notes          string
    Taxonomicclass string
    Slides         []SlideModel `gorm:"foreignKey:TissueRecordID;"`
    Atlases        []AtlasModel `gorm:"many2many:atlas_tissue_records;"`
}

func (TissueRecordModel) TableName() string {
    return "tissue_records"
}
