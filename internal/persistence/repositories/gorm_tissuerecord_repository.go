package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/migration"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var NOT_FOUND_ERROR int = 0
var OK_STATUS int = 1

type GormTissueRecordRepository struct {
	conn string
}

func NewGormTissueRecordRepository() *GormTissueRecordRepository {
	connection := os.Getenv("DB_PATH")
	new_repository := GormTissueRecordRepository{conn: connection}
	return &new_repository
}

func (repo *GormTissueRecordRepository) Save(tr *tissuerecord.TissueRecord) uint {
	db, err := gorm.Open(sqlite.Open(repo.conn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	slides_models := []migration.SlideModel{}
	for _, slide := range tr.Slides {
		// create a staining object
		staining := &migration.StainingModel{
			Name: slide.Staining.Name,
		}

		new_slide := &migration.SlideModel{
			Name:     slide.Name,
			Url:      slide.Img.Url,
			Staining: *staining,
		}
		slides_models = append(slides_models, *new_slide)
	}

	new_tissue_record_model := &migration.TissueRecordModel{
		Name:           tr.Name,
		Notes:          tr.Notes,
		Taxonomicclass: tr.Taxonomicclass,
		Slides:         slides_models,
	}

	db.Create(new_tissue_record_model)
	return new_tissue_record_model.ID
}

func (repo *GormTissueRecordRepository) Delete(id uint) {
	db, err := gorm.Open(sqlite.Open(repo.conn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.Select("slides").Delete(&migration.TissueRecordModel{ID: id})
}

func (repo *GormTissueRecordRepository) Retrieve(id uint) (tissuerecord.TissueRecord, int) {
	db, err := gorm.Open(sqlite.Open(repo.conn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	tissuerecord_found := migration.TissueRecordModel{}

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
	db, err := gorm.Open(sqlite.Open(repo.conn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	slides_models := []migration.SlideModel{}
	for _, slide := range tr.Slides {
		new_slide := &migration.SlideModel{
			Name: slide.Name,
		}
		slides_models = append(slides_models, *new_slide)
	}

	new_tissue_record_model := &migration.TissueRecordModel{
		ID:             id,
		Name:           tr.Name,
		Notes:          tr.Notes,
		Taxonomicclass: tr.Taxonomicclass,
		Slides:         slides_models,
	}

	db.Save(new_tissue_record_model)
}
