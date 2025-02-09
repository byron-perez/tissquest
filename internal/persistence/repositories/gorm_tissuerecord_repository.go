package repositories

import (
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SlideModel struct {
	gorm.Model
	ID             uint `gorm:"primaryKey"`
	Name           string
	TissueRecordID uint
}

type TissueRecordModel struct {
	gorm.Model
	ID             uint `gorm:"primaryKey"`
	Name           string
	Notes          string
	Taxonomicclass string
	Slides         []SlideModel `gorm:"foreignKey:TissueRecordID"`
}

type GormTissueRecordRepository struct {
}

func NewGormTissueRecordRepository() *GormTissueRecordRepository {
	return &GormTissueRecordRepository{}
}

func (repo *GormTissueRecordRepository) Save(tr *tissuerecord.TissueRecord) bool {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&TissueRecordModel{}, &SlideModel{})

	slides_models := []SlideModel{}
	for _, slide := range tr.Slides {
		new_slide := &SlideModel{
			Name: slide.Name,
		}
		slides_models = append(slides_models, *new_slide)
	}

	db.Create(&TissueRecordModel{
		Name:           tr.Name,
		Notes:          tr.Notes,
		Taxonomicclass: tr.Taxonomicclass,
		Slides:         slides_models,
	})
	return true
}

func (repo *GormTissueRecordRepository) Delete(id string) {
}

func (repo *GormTissueRecordRepository) Retrieve(id string) tissuerecord.TissueRecord {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&TissueRecordModel{}, &SlideModel{})
	tissuerecord_found := TissueRecordModel{}
	db.First(&tissuerecord_found, id)

	slides := []slide.Slide{}
	for _, slide_model := range tissuerecord_found.Slides {
		new_slide := &slide.Slide{
			Name: slide_model.Name,
		}
		slides = append(slides, *new_slide)
	}

	mapped_tissue_record := tissuerecord.TissueRecord{
		Name:           tissuerecord_found.Name,
		Notes:          tissuerecord_found.Notes,
		Taxonomicclass: tissuerecord_found.Taxonomicclass,
		Slides:         slides,
	}
	return mapped_tissue_record
}

func (repo *GormTissueRecordRepository) Update(id string, tr *tissuerecord.TissueRecord) {
}
