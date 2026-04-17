package repositories

import (
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/persistence/migration"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GormSlideRepository struct{}

func NewGormSlideRepository() *GormSlideRepository {
	return &GormSlideRepository{}
}

func (repo *GormSlideRepository) getDB() (*gorm.DB, error) {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "postgres" || dbType == "postgresql" {
		return gorm.Open(postgres.Open(buildDSN()), &gorm.Config{})
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "tissquest.db"
	}
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
}

func toSlideModel(sl *slide.Slide) migration.SlideModel {
	return migration.SlideModel{
		Name:          sl.Name,
		Url:           sl.Url,
		TissueRecordID: sl.TissueRecordID,
		Magnification: sl.Magnification,
		Preparation: migration.PreparationModel{
			Staining:        sl.Preparation.Staining,
			InclusionMethod: sl.Preparation.InclusionMethod,
			Reagents:        sl.Preparation.Reagents,
			Protocol:        sl.Preparation.Protocol,
			Notes:           sl.Preparation.Notes,
		},
	}
}

func fromSlideModel(m migration.SlideModel) slide.Slide {
	return slide.Slide{
		ID:             m.ID,
		TissueRecordID: m.TissueRecordID,
		Name:           m.Name,
		Url:            m.Url,
		Magnification:  m.Magnification,
		Preparation: slide.Preparation{
			Staining:        m.Preparation.Staining,
			InclusionMethod: m.Preparation.InclusionMethod,
			Reagents:        m.Preparation.Reagents,
			Protocol:        m.Preparation.Protocol,
			Notes:           m.Preparation.Notes,
		},
	}
}

func (repo *GormSlideRepository) Save(sl *slide.Slide) (uint, error) {
	db, err := repo.getDB()
	if err != nil {
		return 0, err
	}

	model := toSlideModel(sl)
	if err := db.Create(&model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (repo *GormSlideRepository) GetByID(id uint) (*slide.Slide, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var model migration.SlideModel
	if err := db.Preload("Preparation").First(&model, id).Error; err != nil {
		return nil, err
	}

	s := fromSlideModel(model)
	return &s, nil
}

func (repo *GormSlideRepository) Update(id uint, sl *slide.Slide) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	// Update preparation first
	if err := db.Model(&migration.PreparationModel{}).
		Where("id = (SELECT preparation_id FROM slides WHERE id = ?)", id).
		Updates(migration.PreparationModel{
			Staining:        sl.Preparation.Staining,
			InclusionMethod: sl.Preparation.InclusionMethod,
			Reagents:        sl.Preparation.Reagents,
			Protocol:        sl.Preparation.Protocol,
			Notes:           sl.Preparation.Notes,
		}).Error; err != nil {
		return err
	}

	return db.Model(&migration.SlideModel{}).Where("id = ?", id).Updates(migration.SlideModel{
		Name:          sl.Name,
		Url:           sl.Url,
		TissueRecordID: sl.TissueRecordID,
		Magnification: sl.Magnification,
	}).Error
}

func (repo *GormSlideRepository) Delete(id uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	return db.Delete(&migration.SlideModel{}, id).Error
}

func (repo *GormSlideRepository) ListByTissueRecord(tissueRecordID uint) ([]slide.Slide, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var models []migration.SlideModel
	if err := db.Preload("Preparation").Where("tissue_record_id = ?", tissueRecordID).Find(&models).Error; err != nil {
		return nil, err
	}

	slides := make([]slide.Slide, len(models))
	for i, m := range models {
		slides[i] = fromSlideModel(m)
	}
	return slides, nil
}
