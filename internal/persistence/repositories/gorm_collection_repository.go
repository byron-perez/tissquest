package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/collection"
	"mcba/tissquest/internal/persistence/migration"

	"gorm.io/gorm"
)

type GormCollectionRepository struct {
	getDB func() (*gorm.DB, error)
}

func NewGormCollectionRepository() *GormCollectionRepository {
	return &GormCollectionRepository{
		getDB: func() (*gorm.DB, error) {
			return GetDB()
		},
	}
}

// NewGormCollectionRepositoryWithDB creates a repository using a provided DB — for testing.
func NewGormCollectionRepositoryWithDB(db *gorm.DB) *GormCollectionRepository {
	return &GormCollectionRepository{
		getDB: func() (*gorm.DB, error) {
			return db, nil
		},
	}
}

func (r *GormCollectionRepository) Save(c *collection.Collection) (uint, error) {
	db, err := r.getDB()
	if err != nil {
		return 0, err
	}
	model := migration.CollectionModel{
		Name:        c.Name,
		Description: c.Description,
		Goals:       c.Goals,
		Type:        string(c.Type),
		Authors:     c.Authors,
	}
	if err := db.Create(&model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *GormCollectionRepository) Retrieve(id uint) (*collection.Collection, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, err
	}
	var model migration.CollectionModel
	if err := db.
		Preload("Sections", "parent_id IS NULL", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Sections.Assignments", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Sections.Assignments.TissueRecord").
		Preload("Sections.Subsections", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Sections.Subsections.Assignments", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Sections.Subsections.Assignments.TissueRecord").
		First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, collection.ErrNotFound
		}
		return nil, err
	}
	return mapToCollection(model), nil
}

func (r *GormCollectionRepository) Update(id uint, c *collection.Collection) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}
	return db.Model(&migration.CollectionModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        c.Name,
		"description": c.Description,
		"goals":       c.Goals,
		"type":        string(c.Type),
		"authors":     c.Authors,
	}).Error
}

func (r *GormCollectionRepository) Delete(id uint) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}
	// Find all sections for this collection
	var sections []migration.CollectionSectionModel
	db.Where("collection_id = ?", id).Find(&sections)
	for _, s := range sections {
		db.Where("section_id = ?", s.ID).Delete(&migration.CollectionSectionAssignmentModel{})
	}
	db.Where("collection_id = ?", id).Delete(&migration.CollectionSectionModel{})
	return db.Delete(&migration.CollectionModel{}, id).Error
}

func (r *GormCollectionRepository) List() ([]collection.Collection, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, err
	}
	var models []migration.CollectionModel
	if err := db.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]collection.Collection, len(models))
	for i, m := range models {
		result[i] = *mapToCollection(m)
	}
	return result, nil
}

func (r *GormCollectionRepository) CreateSection(s *collection.Section) (uint, error) {
	db, err := r.getDB()
	if err != nil {
		return 0, err
	}
	var count int64
	q := db.Model(&migration.CollectionSectionModel{}).Where("collection_id = ?", s.CollectionID)
	if s.ParentID != nil {
		q = db.Model(&migration.CollectionSectionModel{}).Where("parent_id = ?", *s.ParentID)
	} else {
		q = db.Model(&migration.CollectionSectionModel{}).Where("collection_id = ? AND parent_id IS NULL", s.CollectionID)
	}
	q.Count(&count)

	model := migration.CollectionSectionModel{
		CollectionID: s.CollectionID,
		ParentID:     s.ParentID,
		Name:         s.Name,
		Position:     int(count) + 1,
	}
	if err := db.Create(&model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *GormCollectionRepository) UpdateSection(id uint, s *collection.Section) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}
	return db.Model(&migration.CollectionSectionModel{}).Where("id = ?", id).Update("name", s.Name).Error
}

func (r *GormCollectionRepository) DeleteSection(id uint) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}
	db.Where("section_id = ?", id).Delete(&migration.CollectionSectionAssignmentModel{})
	// Also delete subsections and their assignments
	var subsections []migration.CollectionSectionModel
	db.Where("parent_id = ?", id).Find(&subsections)
	for _, sub := range subsections {
		db.Where("section_id = ?", sub.ID).Delete(&migration.CollectionSectionAssignmentModel{})
	}
	db.Where("parent_id = ?", id).Delete(&migration.CollectionSectionModel{})
	return db.Delete(&migration.CollectionSectionModel{}, id).Error
}

func (r *GormCollectionRepository) ReorderSections(collectionID uint, positions map[uint]int) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}
	for sectionID, pos := range positions {
		if err := db.Model(&migration.CollectionSectionModel{}).
			Where("id = ? AND collection_id = ?", sectionID, collectionID).
			Update("position", pos).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *GormCollectionRepository) CreateAssignment(a *collection.SectionAssignment) (uint, error) {
	db, err := r.getDB()
	if err != nil {
		return 0, err
	}
	// Check for duplicate
	var existing migration.CollectionSectionAssignmentModel
	if err := db.Where("section_id = ? AND tissue_record_id = ?", a.SectionID, a.TissueRecordID).
		First(&existing).Error; err == nil {
		return 0, collection.ErrDuplicateAssignment
	}

	var count int64
	db.Model(&migration.CollectionSectionAssignmentModel{}).Where("section_id = ?", a.SectionID).Count(&count)

	model := migration.CollectionSectionAssignmentModel{
		SectionID:      a.SectionID,
		TissueRecordID: a.TissueRecordID,
		Position:       int(count) + 1,
	}
	if err := db.Create(&model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *GormCollectionRepository) DeleteAssignment(id uint) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}
	// Get the assignment to know its section
	var assignment migration.CollectionSectionAssignmentModel
	if err := db.First(&assignment, id).Error; err != nil {
		return err
	}
	sectionID := assignment.SectionID

	if err := db.Delete(&migration.CollectionSectionAssignmentModel{}, id).Error; err != nil {
		return err
	}

	// Resequence remaining assignments in the section
	var remaining []migration.CollectionSectionAssignmentModel
	db.Where("section_id = ?", sectionID).Order("position ASC").Find(&remaining)
	for i, a := range remaining {
		db.Model(&migration.CollectionSectionAssignmentModel{}).Where("id = ?", a.ID).Update("position", i+1)
	}
	return nil
}

func (r *GormCollectionRepository) ReorderAssignments(sectionID uint, positions map[uint]int) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}
	for assignID, pos := range positions {
		if err := db.Model(&migration.CollectionSectionAssignmentModel{}).
			Where("id = ? AND section_id = ?", assignID, sectionID).
			Update("position", pos).Error; err != nil {
			return err
		}
	}
	return nil
}

// mapToCollection converts a CollectionModel to a domain Collection.
func mapToCollection(m migration.CollectionModel) *collection.Collection {
	collType := collection.CollectionType(m.Type)
	if collType == "" {
		collType = collection.CollectionTypeAtlas
	}

	sections := make([]collection.Section, len(m.Sections))
	for i, s := range m.Sections {
		sections[i] = mapToSection(s)
	}

	return &collection.Collection{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Goals:       m.Goals,
		Type:        collType,
		Authors:     m.Authors,
		Sections:    sections,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func mapToSection(m migration.CollectionSectionModel) collection.Section {
	assignments := make([]collection.SectionAssignment, len(m.Assignments))
	for i, a := range m.Assignments {
		name := ""
		if a.TissueRecord.ID != 0 {
			name = a.TissueRecord.Name
		}
		assignments[i] = collection.SectionAssignment{
			ID:               a.ID,
			SectionID:        a.SectionID,
			TissueRecordID:   a.TissueRecordID,
			TissueRecordName: name,
			Position:         a.Position,
		}
	}
	subsections := make([]collection.Section, len(m.Subsections))
	for i, sub := range m.Subsections {
		subsections[i] = mapToSection(sub)
	}
	return collection.Section{
		ID:           m.ID,
		CollectionID: m.CollectionID,
		ParentID:     m.ParentID,
		Name:         m.Name,
		Position:     m.Position,
		Assignments:  assignments,
		Subsections:  subsections,
	}
}
