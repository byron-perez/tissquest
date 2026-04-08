package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/migration"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresTissueRecordRepository struct {
	dsn string
}

func NewPostgresTissueRecordRepository() *PostgresTissueRecordRepository {
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	sslmode := os.Getenv("DATABASE_SSLMODE")

	if sslmode == "" {
		sslmode = "require"
	}

	dsn := "host=" + host +
		" port=" + port +
		" user=" + user +
		" password=" + password +
		" dbname=" + dbname +
		" sslmode=" + sslmode

	return &PostgresTissueRecordRepository{dsn: dsn}
}

func (repo *PostgresTissueRecordRepository) getDB() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(repo.dsn), &gorm.Config{})
}

func (repo *PostgresTissueRecordRepository) Save(tr *tissuerecord.TissueRecord) uint {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}

	slidesModels := []migration.SlideModel{}
	for _, slideItem := range tr.Slides {
		staining := migration.StainingModel{Name: slideItem.Staining.Name}
		newSlide := migration.SlideModel{
			Name:          slideItem.Name,
			Url:           slideItem.Img.Url,
			Staining:      staining,
			Magnification: slideItem.Magnification,
		}
		slidesModels = append(slidesModels, newSlide)
	}

	newTissueRecordModel := &migration.TissueRecordModel{
		Name:           tr.Name,
		Notes:          tr.Notes,
		Taxonomicclass: tr.Taxonomicclass,
		Slides:         slidesModels,
	}

	db.Create(newTissueRecordModel)
	return newTissueRecordModel.ID
}

func (repo *PostgresTissueRecordRepository) Delete(id uint) {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}
	db.Select("slides").Delete(&migration.TissueRecordModel{ID: id})
}

func (repo *PostgresTissueRecordRepository) Retrieve(id uint) (tissuerecord.TissueRecord, int) {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}

	tissueRecordFound := migration.TissueRecordModel{}
	notFoundError := db.Preload("Slides.Staining").First(&tissueRecordFound, id).Error

	if errors.Is(notFoundError, gorm.ErrRecordNotFound) {
		return tissuerecord.TissueRecord{}, NOT_FOUND_ERROR
	}

	slides := []slide.Slide{}
	for _, slideModel := range tissueRecordFound.Slides {
		slides = append(slides, slide.Slide{
			Name:          slideModel.Name,
			Magnification: slideModel.Magnification,
		})
	}

	return tissuerecord.TissueRecord{
		Name:           tissueRecordFound.Name,
		Notes:          tissueRecordFound.Notes,
		Taxonomicclass: tissueRecordFound.Taxonomicclass,
		Slides:         slides,
	}, OK_STATUS
}

func (repo *PostgresTissueRecordRepository) Update(id uint, tr *tissuerecord.TissueRecord) {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}

	db.Model(&migration.TissueRecordModel{}).Where("id = ?", id).Updates(migration.TissueRecordModel{
		Name:           tr.Name,
		Notes:          tr.Notes,
		Taxonomicclass: tr.Taxonomicclass,
	})
}

func (repo *PostgresTissueRecordRepository) List(page, limit int) ([]tissuerecord.TissueRecord, int64, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, 0, err
	}

	var count int64
	db.Model(&migration.TissueRecordModel{}).Count(&count)

	var tissueRecordModels []migration.TissueRecordModel
	offset := (page - 1) * limit
	result := db.Preload("Slides.Staining").Offset(offset).Limit(limit).Find(&tissueRecordModels)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	records := make([]tissuerecord.TissueRecord, len(tissueRecordModels))
	for i, model := range tissueRecordModels {
		slides := []slide.Slide{}
		for _, slideModel := range model.Slides {
			slides = append(slides, slide.Slide{
				Name:          slideModel.Name,
				Magnification: slideModel.Magnification,
			})
		}
		records[i] = tissuerecord.TissueRecord{
			Name:           model.Name,
			Notes:          model.Notes,
			Taxonomicclass: model.Taxonomicclass,
			Slides:         slides,
		}
	}

	return records, count, nil
}
