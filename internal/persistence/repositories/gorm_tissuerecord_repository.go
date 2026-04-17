package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/migration"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormTissueRecordRepository struct {
	dsn string
}

func NewGormTissueRecordRepository() *GormTissueRecordRepository {
	dsn := buildDSN()
	return &GormTissueRecordRepository{dsn: dsn}
}

func (repo *GormTissueRecordRepository) getDB() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(repo.dsn), &gorm.Config{})
}

func (repo *GormTissueRecordRepository) Save(tr *tissuerecord.TissueRecord) uint {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}

	slideModels := make([]migration.SlideModel, len(tr.Slides))
	for i, s := range tr.Slides {
		slideModels[i] = migration.SlideModel{
			Name:          s.Name,
			Url:           s.Url,
			Magnification: s.Magnification,
			Preparation: migration.PreparationModel{
				Staining:        s.Preparation.Staining,
				InclusionMethod: s.Preparation.InclusionMethod,
				Reagents:        s.Preparation.Reagents,
				Protocol:        s.Preparation.Protocol,
				Notes:           s.Preparation.Notes,
			},
		}
	}

	model := &migration.TissueRecordModel{
		Name:    tr.Name,
		Notes:   tr.Notes,
		TaxonID: tr.TaxonID,
		Slides:  slideModels,
	}
	db.Create(model)
	return model.ID
}

func (repo *GormTissueRecordRepository) Delete(id uint) {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}
	db.Select("Slides").Delete(&migration.TissueRecordModel{Model: gorm.Model{ID: id}})
}

func (repo *GormTissueRecordRepository) Retrieve(id uint) (tissuerecord.TissueRecord, int) {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}

	var model migration.TissueRecordModel
	if err := db.Preload("Slides.Preparation").Preload("Taxon").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tissuerecord.TissueRecord{}, NOT_FOUND_ERROR
		}
		panic(err)
	}

	return mapToTissueRecord(model), OK_STATUS
}

func (repo *GormTissueRecordRepository) Update(id uint, tr *tissuerecord.TissueRecord) {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}

	slideModels := make([]migration.SlideModel, len(tr.Slides))
	for i, s := range tr.Slides {
		slideModels[i] = migration.SlideModel{
			Name:          s.Name,
			Url:           s.Url,
			Magnification: s.Magnification,
			Preparation: migration.PreparationModel{
				Staining:        s.Preparation.Staining,
				InclusionMethod: s.Preparation.InclusionMethod,
				Reagents:        s.Preparation.Reagents,
				Protocol:        s.Preparation.Protocol,
				Notes:           s.Preparation.Notes,
			},
		}
	}

	db.Save(&migration.TissueRecordModel{
		Model:   gorm.Model{ID: id},
		Name:    tr.Name,
		Notes:   tr.Notes,
		TaxonID: tr.TaxonID,
		Slides:  slideModels,
	})
}

func (repo *GormTissueRecordRepository) List(page, limit int) ([]tissuerecord.TissueRecord, int64, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := db.Model(&migration.TissueRecordModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []migration.TissueRecordModel
	offset := (page - 1) * limit
	if err := db.Preload("Slides.Preparation").Preload("Taxon").Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	records := make([]tissuerecord.TissueRecord, len(models))
	for i, m := range models {
		records[i] = mapToTissueRecord(m)
	}
	return records, total, nil
}

func mapToTissueRecord(m migration.TissueRecordModel) tissuerecord.TissueRecord {
	slides := make([]slide.Slide, len(m.Slides))
	for i, s := range m.Slides {
		slides[i] = slide.Slide{
			Name:          s.Name,
			Url:           s.Url,
			Magnification: s.Magnification,
			Preparation: slide.Preparation{
				Staining:        s.Preparation.Staining,
				InclusionMethod: s.Preparation.InclusionMethod,
				Reagents:        s.Preparation.Reagents,
				Protocol:        s.Preparation.Protocol,
				Notes:           s.Preparation.Notes,
			},
		}
	}
	tr := tissuerecord.TissueRecord{
		ID:      m.ID,
		Name:    m.Name,
		Notes:   m.Notes,
		TaxonID: m.TaxonID,
		Slides:  slides,
	}
	if m.Taxon.ID != 0 {
		tr.Taxon = modelToTaxonDeep(m.Taxon)
	}
	return tr
}

func buildDSN() string {
	if dbType := os.Getenv("DB_TYPE"); dbType == "postgres" || dbType == "postgresql" {
		return "host=" + os.Getenv("DATABASE_HOST") +
			" user=" + os.Getenv("DATABASE_USER") +
			" password=" + os.Getenv("DATABASE_PASSWORD") +
			" dbname=" + os.Getenv("DATABASE_NAME") +
			" port=" + os.Getenv("DATABASE_PORT") +
			" sslmode=require TimeZone=UTC"
	}
	return os.Getenv("DB_PATH")
}

func (repo *GormTissueRecordRepository) AddAtlas(trID, atlasID uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}
	model := migration.TissueRecordModel{}
	model.ID = trID
	atlasModel := migration.AtlasModel{}
	atlasModel.ID = atlasID
	return db.Model(&model).Association("Atlases").Append(&atlasModel)
}

func (repo *GormTissueRecordRepository) RemoveAtlas(trID, atlasID uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}
	model := migration.TissueRecordModel{}
	model.ID = trID
	atlasModel := migration.AtlasModel{}
	atlasModel.ID = atlasID
	return db.Model(&model).Association("Atlases").Delete(&atlasModel)
}

func (repo *GormTissueRecordRepository) ListAtlases(trID uint) ([]atlas.Atlas, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}
	model := migration.TissueRecordModel{}
	model.ID = trID
	var atlasModels []migration.AtlasModel
	if err := db.Model(&model).Association("Atlases").Find(&atlasModels); err != nil {
		return nil, err
	}
	result := make([]atlas.Atlas, len(atlasModels))
	for i, a := range atlasModels {
		result[i] = atlas.Atlas{
			ID:          a.ID,
			Name:        a.Name,
			Description: a.Description,
			Category:    a.Category,
		}
	}
	return result, nil
}

func (repo *GormTissueRecordRepository) AddCategory(trID, catID uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}
	model := migration.TissueRecordModel{}
	model.ID = trID
	catModel := migration.CategoryModel{}
	catModel.ID = catID
	return db.Model(&model).Association("Categories").Append(&catModel)
}

func (repo *GormTissueRecordRepository) RemoveCategory(trID, catID uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}
	model := migration.TissueRecordModel{}
	model.ID = trID
	catModel := migration.CategoryModel{}
	catModel.ID = catID
	return db.Model(&model).Association("Categories").Delete(&catModel)
}

func (repo *GormTissueRecordRepository) ListCategories(trID uint) ([]category.Category, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}
	model := migration.TissueRecordModel{}
	model.ID = trID
	var catModels []migration.CategoryModel
	if err := db.Model(&model).Association("Categories").Find(&catModels); err != nil {
		return nil, err
	}
	result := make([]category.Category, len(catModels))
	for i, c := range catModels {
		result[i] = category.Category{
			ID:          c.ID,
			Name:        c.Name,
			Type:        category.CategoryType(c.Type),
			Description: c.Description,
			ParentID:    c.ParentID,
		}
	}
	return result, nil
}
