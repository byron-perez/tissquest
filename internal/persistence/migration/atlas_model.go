package migration

import "gorm.io/gorm"

type AtlasModel struct {
	gorm.Model
	ID          uint `gorm:"primaryKey"`
	Name        string
	Description string
	Category    string
	// Use the column tag to override the default foreign key name
	TissueRecords []TissueRecordModel `gorm:"many2many:atlas_tissue_records;foreignKey:ID;joinForeignKey:AtlasID;references:ID;joinReferences:TissueRecordID"`
}

func (AtlasModel) TableName() string {
	return "atlases"
}
