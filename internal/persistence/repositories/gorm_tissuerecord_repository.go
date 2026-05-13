package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/migration"

	"gorm.io/gorm"
)

type GormTissueRecordRepository struct {
	getDB func() (*gorm.DB, error)
}

func NewGormTissueRecordRepository() *GormTissueRecordRepository {
	return &GormTissueRecordRepository{
		getDB: GetDB,
	}
}

// NewGormTissueRecordRepositoryWithDB creates a repository using a provided DB — for testing.
func NewGormTissueRecordRepositoryWithDB(db *gorm.DB) *GormTissueRecordRepository {
	return &GormTissueRecordRepository{
		getDB: func() (*gorm.DB, error) { return db, nil },
	}
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
			ImageKey:      s.ImageKey,
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
			ImageKey:      s.ImageKey,
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
			ID:            s.ID,
			Name:          s.Name,
			ImageKey:      s.ImageKey,
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

func (repo *GormTissueRecordRepository) Search(q string, categoryIDs []uint, page, limit int) ([]tissuerecord.TissueRecord, int64, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, 0, err
	}

	base := db.Model(&migration.TissueRecordModel{}).
		Joins("LEFT JOIN taxa ON taxa.id = tissue_records.taxon_id")

	if q != "" {
		pattern := "%" + q + "%"
		base = base.Where("tissue_records.name ILIKE ? OR taxa.name ILIKE ?", pattern, pattern)
	}

	if len(categoryIDs) > 0 {
		base = base.
			Joins("JOIN tissue_record_categories ON tissue_record_categories.tissue_record_id = tissue_records.id").
			Where("tissue_record_categories.category_id IN ?", categoryIDs)
	}

	var total int64
	if err := base.Distinct("tissue_records.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []migration.TissueRecordModel
	offset := (page - 1) * limit
	if err := base.
		Distinct("tissue_records.*").
		Preload("Slides.Preparation").
		Preload("Taxon").
		Offset(offset).Limit(limit).
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	if len(models) == 0 {
		return nil, total, nil
	}

	// Collect all slide IDs across all tissue records in one pass
	var allSlideIDs []uint
	for _, m := range models {
		for _, s := range m.Slides {
			allSlideIDs = append(allSlideIDs, s.ID)
		}
	}

	// Fetch all image variants for those slides in a single query
	variantMap := make(map[uint]map[string]string) // slideID → size → url
	if len(allSlideIDs) > 0 {
		var variants []migration.SlideImageVariantModel
		db.Where("slide_id IN ?", allSlideIDs).Find(&variants)
		for _, v := range variants {
			if variantMap[v.SlideID] == nil {
				variantMap[v.SlideID] = make(map[string]string)
			}
			variantMap[v.SlideID][v.Size] = v.Url
		}
	}

	records := make([]tissuerecord.TissueRecord, len(models))
	for i, m := range models {
		tr := mapToTissueRecord(m)
		// Resolve the best available thumbnail across all slides
		for _, s := range m.Slides {
			sizes := variantMap[s.ID]
			url := resolveVariantURL(sizes)
			if url != "" {
				tr.FeaturedImageURL = url
				break
			}
		}
		records[i] = tr
	}
	return records, total, nil
}

// resolveVariantURL picks the best available URL: thumb → original → empty.
func resolveVariantURL(sizes map[string]string) string {
	if sizes == nil {
		return ""
	}
	if url := sizes[string(slide.ImageSizeThumb)]; url != "" {
		return url
	}
	if url := sizes[string(slide.ImageSizePreview)]; url != "" {
		return url
	}
	if url := sizes[string(slide.ImageSizeOriginal)]; url != "" {
		return url
	}
	return ""
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
