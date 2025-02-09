package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var NOT_FOUND_ERROR int = 0
var OK_STATUS int = 1

type Tabler interface {
	TableName() string
}

type SlideModel struct {
	gorm.Model
	ID             uint `gorm:"primaryKey"`
	Name           string
	TissueRecordID uint
}

func (SlideModel) TableName() string {
	return "slides"
}

type TissueRecordModel struct {
	gorm.Model
	ID             uint `gorm:"primaryKey"`
	Name           string
	Notes          string
	Taxonomicclass string
	Slides         []SlideModel `gorm:"foreignKey:TissueRecordID;"`
}

func (TissueRecordModel) TableName() string {
	return "tissue_records"
}

type GormTissueRecordRepository struct {
}

func NewGormTissueRecordRepository() *GormTissueRecordRepository {
	return &GormTissueRecordRepository{}
}

func (repo *GormTissueRecordRepository) Save(tr *tissuerecord.TissueRecord) uint {
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

	new_tissue_record_model := &TissueRecordModel{
		Name:           tr.Name,
		Notes:          tr.Notes,
		Taxonomicclass: tr.Taxonomicclass,
		Slides:         slides_models,
	}

	db.Create(new_tissue_record_model)
	return new_tissue_record_model.ID
}

func (repo *GormTissueRecordRepository) Delete(id uint) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.Select("slides").Delete(&TissueRecordModel{ID: id})
}

func (repo *GormTissueRecordRepository) Retrieve(id uint) (tissuerecord.TissueRecord, int) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	tissuerecord_found := TissueRecordModel{}

	not_found_error := db.First(&tissuerecord_found, id).Error

	if errors.Is(not_found_error, gorm.ErrRecordNotFound) {
		return tissuerecord.TissueRecord{}, NOT_FOUND_ERROR
	}

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
	return mapped_tissue_record, OK_STATUS
}

func (repo *GormTissueRecordRepository) Update(id uint, tr *tissuerecord.TissueRecord) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	slides_models := []SlideModel{}
	for _, slide := range tr.Slides {
		new_slide := &SlideModel{
			Name: slide.Name,
		}
		slides_models = append(slides_models, *new_slide)
	}

	new_tissue_record_model := &TissueRecordModel{
		ID:             id,
		Name:           tr.Name,
		Notes:          tr.Notes,
		Taxonomicclass: tr.Taxonomicclass,
		Slides:         slides_models,
	}

	db.Save(new_tissue_record_model)
}
