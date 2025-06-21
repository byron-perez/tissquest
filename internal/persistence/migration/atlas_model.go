package migration

import "gorm.io/gorm"

type AtlasModel struct {
    gorm.Model
    ID          uint `gorm:"primaryKey"`
    Name        string
    Description string
    Category    string
    TissueRecords []TissueRecordModel `gorm:"many2many:atlas_tissue_records;"`
}