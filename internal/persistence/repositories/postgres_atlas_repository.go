package repositories

import (
	"fmt"
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/persistence/migration"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresAtlasRepository struct {
	dsn string
}

func NewPostgresAtlasRepository() *PostgresAtlasRepository {
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	sslmode := os.Getenv("DATABASE_SSLMODE")

	if sslmode == "" {
		sslmode = "require" // Default to disable if not specified
	}

	// Construct the DSN string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	fmt.Println(dsn)
	new_repository := PostgresAtlasRepository{dsn: dsn}
	return &new_repository
}

func (repo *PostgresAtlasRepository) getDB() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(repo.dsn), &gorm.Config{})
}

func (repo *PostgresAtlasRepository) Save(a *atlas.Atlas) uint {
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

func (repo *PostgresAtlasRepository) Retrieve(id uint) (*atlas.Atlas, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var atlasModel migration.AtlasModel
	result := db.First(&atlasModel, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &atlas.Atlas{
		ID:          atlasModel.ID,
		Name:        atlasModel.Name,
		Description: atlasModel.Description,
		Category:    atlasModel.Category,
	}, nil
}

func (repo *PostgresAtlasRepository) Update(id uint, a *atlas.Atlas) error {
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

func (repo *PostgresAtlasRepository) Delete(id uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	result := db.Delete(&migration.AtlasModel{}, id)
	return result.Error
}

func (repo *PostgresAtlasRepository) List() ([]atlas.Atlas, error) {
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

// FindByName searches for atlases by name
func (repo *PostgresAtlasRepository) FindByName(name string) ([]atlas.Atlas, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var atlasModels []migration.AtlasModel
	// Using ILIKE for case-insensitive search in PostgreSQL
	result := db.Where("name ILIKE ?", "%"+name+"%").Find(&atlasModels)
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

// FindByCategory searches for atlases by category
func (repo *PostgresAtlasRepository) FindByCategory(category string) ([]atlas.Atlas, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var atlasModels []migration.AtlasModel
	result := db.Where("category = ?", category).Find(&atlasModels)
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

// ListWithPagination retrieves atlases with pagination
func (repo *PostgresAtlasRepository) ListWithPagination(page, pageSize int) ([]atlas.Atlas, int64, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, 0, err
	}

	var count int64
	db.Model(&migration.AtlasModel{}).Count(&count)

	var atlasModels []migration.AtlasModel
	offset := (page - 1) * pageSize
	result := db.Offset(offset).Limit(pageSize).Find(&atlasModels)
	if result.Error != nil {
		return nil, 0, result.Error
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
	return atlases, count, nil
}
