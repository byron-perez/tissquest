package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/persistence/migration"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GormCategoryRepository struct {
	conn string
}

func NewGormCategoryRepository() *GormCategoryRepository {
	connection := os.Getenv("DB_PATH")
	if connection == "" {
		connection = "tissquest.db"
	}
	return &GormCategoryRepository{conn: connection}
}

func (repo *GormCategoryRepository) getDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(repo.conn), &gorm.Config{})
}

func (repo *GormCategoryRepository) Save(c *category.Category) uint {
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

func (repo *GormCategoryRepository) Retrieve(id uint) (*category.Category, error) {
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

func (repo *GormCategoryRepository) Update(id uint, c *category.Category) error {
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

func (repo *GormCategoryRepository) Delete(id uint) error {
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

func (repo *GormCategoryRepository) List() ([]category.Category, error) {
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

func (repo *GormCategoryRepository) FindByType(categoryType category.CategoryType) ([]category.Category, error) {
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

func (repo *GormCategoryRepository) FindByParent(parentID uint) ([]category.Category, error) {
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

func (repo *GormCategoryRepository) FindRootCategories() ([]category.Category, error) {
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

func mapCategoryModelToDomain(m *migration.CategoryModel) *category.Category {
	if m == nil {
		return nil
	}

	ids := make([]uint, 0, len(m.TissueRecords))
	for _, tr := range m.TissueRecords {
		ids = append(ids, tr.ID)
	}

	return &category.Category{
		ID:              m.ID,
		Name:            m.Name,
		Type:            category.CategoryType(m.Type),
		Description:     m.Description,
		ParentID:        m.ParentID,
		TissueRecordIDs: ids,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func mapCategoryModelsToDomain(models []migration.CategoryModel) []category.Category {
	categories := make([]category.Category, len(models))
	for i, m := range models {
		categories[i] = *mapCategoryModelToDomain(&m)
	}
	return categories
}
