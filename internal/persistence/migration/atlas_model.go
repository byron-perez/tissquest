package migration

import "gorm.io/gorm"

type AtlasModel struct {
	gorm.Model
	Name        string
	Description string
	Category    string
	Categories  []CategoryModel `gorm:"many2many:atlas_categories;"`
	TissueRecords []TissueRecordModel `gorm:"many2many:atlas_tissue_records;joinForeignKey:AtlasID;joinReferences:TissueRecordID"`
}

func (AtlasModel) TableName() string {
	return "atlases"
}
