package migration

import "gorm.io/gorm"

type TissueRecordModel struct {
    gorm.Model
    Name       string
    Notes      string
    TaxonID    *uint
    Taxon      TaxonModel      `gorm:"foreignKey:TaxonID"`
    Slides     []SlideModel    `gorm:"foreignKey:TissueRecordID;"`
    Categories []CategoryModel `gorm:"many2many:tissue_record_categories;joinForeignKey:TissueRecordID;joinReferences:CategoryID"`
    Atlases    []AtlasModel    `gorm:"many2many:atlas_tissue_records;joinForeignKey:TissueRecordID;joinReferences:AtlasID"`
}

func (TissueRecordModel) TableName() string {
    return "tissue_records"
}
