package repositories

import (
	"fmt"
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/persistence/migration"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GormAtlasRepository struct {
	conn string
}

func NewGormAtlasRepository() *GormAtlasRepository {
	connection := os.Getenv("DB_PATH")
	fmt.Println(connection)
	new_repository := GormAtlasRepository{conn: connection}
	return &new_repository
}

func (repo *GormAtlasRepository) getDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(repo.conn), &gorm.Config{})
}

func (repo *GormAtlasRepository) Save(a *atlas.Atlas) uint {
	db, err := repo.getDB()
	if err != nil {
		// Handle error (e.g., log it or panic)
		panic("failed to connect database")
	}

	atlasModel := migration.AtlasModel{
		Name:        a.Name,
		Description: a.Description,
		Category:    a.Category,
	}
	db.Create(&atlasModel)
	return atlasModel.ID
}

func (repo *GormAtlasRepository) Retrieve(id uint) (*atlas.Atlas, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var atlasModel migration.AtlasModel
	if err := db.First(&atlasModel, id).Error; err != nil {
		return nil, err
	}
	return &atlas.Atlas{
		ID:          atlasModel.ID,
		Name:        atlasModel.Name,
		Description: atlasModel.Description,
		Category:    atlasModel.Category,
	}, nil
}

func (repo *GormAtlasRepository) Update(id uint, a *atlas.Atlas) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	result := db.Model(&migration.AtlasModel{}).Where("id = ?", id).Updates(migration.AtlasModel{
		Name:        a.Name,
		Description: a.Description,
		Category:    a.Category,
	})
	return result.Error
}

func (repo *GormAtlasRepository) Delete(id uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	result := db.Delete(&migration.AtlasModel{}, id)
	return result.Error
}

func (repo *GormAtlasRepository) FindByName(name string) ([]atlas.Atlas, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var atlasModels []migration.AtlasModel
	if err := db.Where("name LIKE ?", "%"+name+"%").Find(&atlasModels).Error; err != nil {
		return nil, err
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

func (repo *GormAtlasRepository) ListWithPagination(page, pageSize int) ([]atlas.Atlas, int64, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := db.Model(&migration.AtlasModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var atlasModels []migration.AtlasModel
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&atlasModels).Error; err != nil {
		return nil, 0, err
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
	return atlases, total, nil
}

func (repo *GormAtlasRepository) List() ([]atlas.Atlas, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var atlasModels []migration.AtlasModel
	result := db.Find(&atlasModels)
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
