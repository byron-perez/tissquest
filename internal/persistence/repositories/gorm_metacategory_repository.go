package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/metacategory"
	"mcba/tissquest/internal/persistence/migration"

	"gorm.io/gorm"
)

type GormMetacategoryRepository struct{}

func NewGormMetacategoryRepository() *GormMetacategoryRepository {
	return &GormMetacategoryRepository{}
}

func (repo *GormMetacategoryRepository) getDB() (*gorm.DB, error) {
	return GetDB()
}

func (repo *GormMetacategoryRepository) Save(m *metacategory.Metacategory) (uint, error) {
	db, err := repo.getDB()
	if err != nil {
		return 0, err
	}

	if err := m.Validate(); err != nil {
		return 0, err
	}

	metaCatModel := migration.MetacategoryModel{
		Name:        m.Name,
		Description: m.Description,
		ParentID:    m.ParentID,
	}

	if err := db.Create(&metaCatModel).Error; err != nil {
		return 0, err
	}
	return metaCatModel.ID, nil
}

func (repo *GormMetacategoryRepository) Retrieve(id uint) (*metacategory.Metacategory, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var model migration.MetacategoryModel
	if err := db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, metacategory.ErrNotFound
		}
		return nil, err
	}

	return mapMetacategoryModelToDomain(&model), nil
}

func (repo *GormMetacategoryRepository) Update(id uint, m *metacategory.Metacategory) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	if err := m.Validate(); err != nil {
		return err
	}

	return db.Model(&migration.MetacategoryModel{}).Where("id = ?", id).Updates(migration.MetacategoryModel{
		Name:        m.Name,
		Description: m.Description,
		ParentID:    m.ParentID,
	}).Error
}

func (repo *GormMetacategoryRepository) Delete(id uint) error {
	db, err := repo.getDB()
	if err != nil {
		return err
	}

	result := db.Delete(&migration.MetacategoryModel{}, id)
	if result.RowsAffected == 0 {
		return metacategory.ErrNotFound
	}
	return result.Error
}

func (repo *GormMetacategoryRepository) List() ([]metacategory.Metacategory, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var models []migration.MetacategoryModel
	if err := db.Where("parent_id IS NULL").Find(&models).Error; err != nil {
		return nil, err
	}

	return mapMetacategoryModelsToDomain(models), nil
}

func (repo *GormMetacategoryRepository) FindByParent(parentID uint) ([]metacategory.Metacategory, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var models []migration.MetacategoryModel
	if err := db.Where("parent_id = ?", parentID).Find(&models).Error; err != nil {
		return nil, err
	}

	return mapMetacategoryModelsToDomain(models), nil
}

func (repo *GormMetacategoryRepository) FindRootMetacategories() ([]metacategory.Metacategory, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var models []migration.MetacategoryModel
	if err := db.Where("parent_id IS NULL").Find(&models).Error; err != nil {
		return nil, err
	}

	return mapMetacategoryModelsToDomain(models), nil
}

func (repo *GormMetacategoryRepository) GetFullHierarchy(id uint) (*metacategory.Metacategory, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var model migration.MetacategoryModel
	if err := db.Preload("Parent").Preload("Children").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, metacategory.ErrNotFound
		}
		return nil, err
	}

	return mapMetacategoryModelWithHierarchyToDomain(&model), nil
}

func (repo *GormMetacategoryRepository) ListWithChildren() ([]metacategory.Metacategory, error) {
	db, err := repo.getDB()
	if err != nil {
		return nil, err
	}

	var models []migration.MetacategoryModel
	if err := db.Where("parent_id IS NULL").Preload("Children").Find(&models).Error; err != nil {
		return nil, err
	}

	return mapMetacategoryModelsToDomain(models), nil
}

// Mapping functions

func mapMetacategoryModelToDomain(m *migration.MetacategoryModel) *metacategory.Metacategory {
	if m == nil {
		return nil
	}

	return &metacategory.Metacategory{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		ParentID:    m.ParentID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func mapMetacategoryModelsToDomain(models []migration.MetacategoryModel) []metacategory.Metacategory {
	result := make([]metacategory.Metacategory, len(models))
	for i, m := range models {
		result[i] = *mapMetacategoryModelToDomain(&m)
	}
	return result
}

func mapMetacategoryModelWithHierarchyToDomain(m *migration.MetacategoryModel) *metacategory.Metacategory {
	if m == nil {
		return nil
	}

	result := &metacategory.Metacategory{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		ParentID:    m.ParentID,
		Parent:      mapMetacategoryModelToDomain(m.Parent),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	if m.Children != nil {
		result.Children = mapMetacategoryModelsToDomain(m.Children)
	}

	return result
}
