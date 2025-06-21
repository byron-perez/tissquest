package repositories

import (
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/persistence/migration"

	"gorm.io/gorm"
)

type GormAtlasRepository struct {
	db *gorm.DB
}

func NewGormAtlasRepository(db *gorm.DB) *GormAtlasRepository {
	return &GormAtlasRepository{db: db}
}

func (repo *GormAtlasRepository) Save(a *atlas.Atlas) uint {
	atlasModel := migration.AtlasModel{
		Name:        a.Name,
		Description: a.Description,
		Category:    a.Category,
	}
	repo.db.Create(&atlasModel)
	return atlasModel.ID
}

func (repo *GormAtlasRepository) Retrieve(id uint) (atlas.Atlas, error) {
	var atlasModel migration.AtlasModel
	result := repo.db.First(&atlasModel, id)
	if result.Error != nil {
		return atlas.Atlas{}, result.Error
	}
	return atlas.Atlas{
		ID:          atlasModel.ID,
		Name:        atlasModel.Name,
		Description: atlasModel.Description,
		Category:    atlasModel.Category,
	}, nil
}

func (repo *GormAtlasRepository) Update(id uint, a *atlas.Atlas) error {
	result := repo.db.Model(&migration.AtlasModel{}).Where("id = ?", id).Updates(migration.AtlasModel{
		Name:        a.Name,
		Description: a.Description,
		Category:    a.Category,
	})
	return result.Error
}

func (repo *GormAtlasRepository) Delete(id uint) error {
	result := repo.db.Delete(&migration.AtlasModel{}, id)
	return result.Error
}

func (repo *GormAtlasRepository) List() ([]atlas.Atlas, error) {
	var atlasModels []migration.AtlasModel
	result := repo.db.Find(&atlasModels)
	if result.Error != nil {
		return nil, result.Error
	}

	atlases := make([]atlas.Atlas, len(atlasModels))
	for i, model := range atlasModels {
		atlases[i] = atlas.Atlas{
			ID:          model.ID,
			Name:        model.Name,
			Description: model.Description,
			Category:    model.Category,
		}
	}
	return atlases, nil
}
