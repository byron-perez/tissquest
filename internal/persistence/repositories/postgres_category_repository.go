package repositories

import (
	"fmt"
	"errors"
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/persistence/migration"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresCategoryRepository struct {
	dsn string
}

func NewPostgresCategoryRepository() *PostgresCategoryRepository {
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	sslmode := os.Getenv("DATABASE_SSLMODE")

	if sslmode == "" {
		sslmode = "require"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	return &PostgresCategoryRepository{dsn: dsn}
}

func (repo *PostgresCategoryRepository) getDB() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(repo.dsn), &gorm.Config{})
}

func (repo *PostgresCategoryRepository) Save(c *category.Category) uint {
	db, err := repo.getDB()
	if err != nil {
		panic("failed to connect database")
	}

	categoryModel := migration.CategoryModel{
		Name:        c.Name,
		Type:        string(c.Type),
		Description: c.Description,
		ParentID:    c.ParentID,
	}

	db.Create(&categoryModel)
	return categoryModel.ID
}

func (repo *PostgresCategoryRepository) Retrieve(id uint) (*category.Category, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var categoryModel migration.CategoryModel
	if err := db.Preload("TissueRecords").First(&categoryModel, id).Error; err != nil {
		return nil, err
	}

	return mapCategoryModelToDomain(&categoryModel), nil
}

func (repo *PostgresCategoryRepository) Update(id uint, c *category.Category) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	result := db.Model(&migration.CategoryModel{}).Where("id = ?", id).Updates(migration.CategoryModel{
		Name:        c.Name,
		Type:        string(c.Type),
		Description: c.Description,
		ParentID:    c.ParentID,
	})
	return result.Error
}

func (repo *PostgresCategoryRepository) Delete(id uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	result := db.Delete(&migration.CategoryModel{}, id)
	if result.RowsAffected == 0 {
		return errors.New("category not found")
	}
	return result.Error
}

func (repo *PostgresCategoryRepository) List() ([]category.Category, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var categoryModels []migration.CategoryModel
	if err := db.Preload("TissueRecords").Find(&categoryModels).Error; err != nil {
		return nil, err
	}

	return mapCategoryModelsToDomain(categoryModels), nil
}

func (repo *PostgresCategoryRepository) FindByType(categoryType category.CategoryType) ([]category.Category, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var categoryModels []migration.CategoryModel
	if err := db.Preload("TissueRecords").Where("type = ?", string(categoryType)).Find(&categoryModels).Error; err != nil {
		return nil, err
	}

	return mapCategoryModelsToDomain(categoryModels), nil
}

func (repo *PostgresCategoryRepository) FindByParent(parentID uint) ([]category.Category, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var categoryModels []migration.CategoryModel
	if err := db.Preload("TissueRecords").Where("parent_id = ?", parentID).Find(&categoryModels).Error; err != nil {
		return nil, err
	}

	return mapCategoryModelsToDomain(categoryModels), nil
}

func (repo *PostgresCategoryRepository) FindRootCategories() ([]category.Category, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var categoryModels []migration.CategoryModel
	if err := db.Preload("TissueRecords").Where("parent_id IS NULL").Find(&categoryModels).Error; err != nil {
		return nil, err
	}

	return mapCategoryModelsToDomain(categoryModels), nil
}
