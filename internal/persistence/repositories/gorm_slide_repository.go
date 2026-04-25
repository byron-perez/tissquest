package repositories

import (
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/persistence/migration"

	"gorm.io/gorm"
)

type GormSlideRepository struct{}

func NewGormSlideRepository() *GormSlideRepository {
	return &GormSlideRepository{}
}

func (repo *GormSlideRepository) getDB() (*gorm.DB, error) {
	return GetDB()
}

func toSlideModel(sl *slide.Slide) migration.SlideModel {
	return migration.SlideModel{
		Name:           sl.Name,
		ImageKey:       sl.ImageKey,
		TissueRecordID: sl.TissueRecordID,
		Magnification:  sl.Magnification,
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
		ImageKey:       m.ImageKey,
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
		Name:           sl.Name,
		ImageKey:       sl.ImageKey,
		TissueRecordID: sl.TissueRecordID,
		Magnification:  sl.Magnification,
	}).Error
}

// SetImageVariant stores or updates the URL for a specific size variant.
// This is called by the Lambda callback — the domain never touches this.
func (repo *GormSlideRepository) SetImageVariant(slideID uint, size slide.ImageSize, url string) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}
	variant := migration.SlideImageVariantModel{
		SlideID: slideID,
		Size:    string(size),
		Url:     url,
	}
	return db.Where(migration.SlideImageVariantModel{SlideID: slideID, Size: string(size)}).
		Assign(migration.SlideImageVariantModel{Url: url}).
		FirstOrCreate(&variant).Error
}

func (repo *GormSlideRepository) Delete(id uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}
	// Remove image variants first
	db.Where("slide_id = ?", id).Delete(&migration.SlideImageVariantModel{})
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

// ListDisplayByTissueRecord returns slides with the best available image URL
// for the preferred size, falling back to original if not available.
func (repo *GormSlideRepository) ListDisplayByTissueRecord(tissueRecordID uint, preferredSize slide.ImageSize) ([]slide.DisplaySlide, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var models []migration.SlideModel
	if err := db.Preload("Preparation").Where("tissue_record_id = ?", tissueRecordID).Find(&models).Error; err != nil {
		return nil, err
	}

	if len(models) == 0 {
		return nil, nil
	}

	// Fetch all variants for these slides in one query
	ids := make([]uint, len(models))
	for i, m := range models {
		ids[i] = m.ID
	}
	var variants []migration.SlideImageVariantModel
	db.Where("slide_id IN ?", ids).Find(&variants)

	// Index variants by slideID → size → url
	variantMap := make(map[uint]map[string]string)
	for _, v := range variants {
		if variantMap[v.SlideID] == nil {
			variantMap[v.SlideID] = make(map[string]string)
		}
		variantMap[v.SlideID][v.Size] = v.Url
	}

	result := make([]slide.DisplaySlide, len(models))
	for i, m := range models {
		sl := fromSlideModel(m)
		sizes := variantMap[m.ID]
		imageUrl := resolveImageUrl(sizes, preferredSize)
		result[i] = slide.DisplaySlide{Slide: sl, ImageUrl: imageUrl}
	}
	return result, nil
}

// resolveImageUrl picks the best available URL: preferred size → original → empty.
func resolveImageUrl(sizes map[string]string, preferred slide.ImageSize) string {
	if sizes == nil {
		return ""
	}
	if url, ok := sizes[string(preferred)]; ok && url != "" {
		return url
	}
	if url, ok := sizes[string(slide.ImageSizeOriginal)]; ok && url != "" {
		return url
	}
	return ""
}
