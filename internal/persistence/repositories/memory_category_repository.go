package repositories

import (
	"errors"
	"mcba/tissquest/internal/core/category"
	"sync"
	"time"
)

type MemoryCategoryRepository struct {
	categories map[uint]*category.Category
	nextID     uint
	mutex      sync.RWMutex
}

func NewMemoryCategoryRepository() *MemoryCategoryRepository {
	repo := &MemoryCategoryRepository{
		categories: make(map[uint]*category.Category),
		nextID:     1,
	}
	repo.populateTestData()
	return repo
}

func (r *MemoryCategoryRepository) Save(c *category.Category) uint {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	c.ID = r.nextID
	r.nextID++
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	
	r.categories[c.ID] = c
	return c.ID
}

func (r *MemoryCategoryRepository) Retrieve(id uint) (*category.Category, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	cat, exists := r.categories[id]
	if !exists {
		return nil, errors.New("category not found")
	}
	return cat, nil
}

func (r *MemoryCategoryRepository) Update(id uint, c *category.Category) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.categories[id]; !exists {
		return errors.New("category not found")
	}
	
	c.ID = id
	c.UpdatedAt = time.Now()
	r.categories[id] = c
	return nil
}

func (r *MemoryCategoryRepository) Delete(id uint) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.categories[id]; !exists {
		return errors.New("category not found")
	}
	
	delete(r.categories, id)
	return nil
}

func (r *MemoryCategoryRepository) List() ([]category.Category, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make([]category.Category, 0, len(r.categories))
	for _, cat := range r.categories {
		result = append(result, *cat)
	}
	return result, nil
}

func (r *MemoryCategoryRepository) FindByType(categoryType category.CategoryType) ([]category.Category, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var result []category.Category
	for _, cat := range r.categories {
		if cat.Type == categoryType {
			result = append(result, *cat)
		}
	}
	return result, nil
}

func (r *MemoryCategoryRepository) FindByParent(parentID uint) ([]category.Category, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var result []category.Category
	for _, cat := range r.categories {
		if cat.ParentID != nil && *cat.ParentID == parentID {
			result = append(result, *cat)
		}
	}
	return result, nil
}

func (r *MemoryCategoryRepository) FindRootCategories() ([]category.Category, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var result []category.Category
	for _, cat := range r.categories {
		if cat.ParentID == nil {
			result = append(result, *cat)
		}
	}
	return result, nil
}

func (r *MemoryCategoryRepository) populateTestData() {
	now := time.Now()
	
	// Species categories
	human := &category.Category{ID: 1, Name: "Human", Type: category.CategorySpecies, Description: "Homo sapiens", CreatedAt: now, UpdatedAt: now}
	mouse := &category.Category{ID: 2, Name: "Mouse", Type: category.CategorySpecies, Description: "Mus musculus", CreatedAt: now, UpdatedAt: now}
	rat := &category.Category{ID: 3, Name: "Rat", Type: category.CategorySpecies, Description: "Rattus norvegicus", CreatedAt: now, UpdatedAt: now}
	
	// Organ categories
	heart := &category.Category{ID: 4, Name: "Heart", Type: category.CategoryOrgan, Description: "Cardiac organ", CreatedAt: now, UpdatedAt: now}
	liver := &category.Category{ID: 5, Name: "Liver", Type: category.CategoryOrgan, Description: "Hepatic organ", CreatedAt: now, UpdatedAt: now}
	brain := &category.Category{ID: 6, Name: "Brain", Type: category.CategoryOrgan, Description: "Central nervous system", CreatedAt: now, UpdatedAt: now}
	kidney := &category.Category{ID: 7, Name: "Kidney", Type: category.CategoryOrgan, Description: "Renal organ", CreatedAt: now, UpdatedAt: now}
	
	// Tissue categories with parent relationships
	heartID := uint(4)
	liverID := uint(5)
	brainID := uint(6)
	
	cardiacMuscle := &category.Category{ID: 8, Name: "Cardiac Muscle", Type: category.CategoryTissue, ParentID: &heartID, Description: "Heart muscle tissue", CreatedAt: now, UpdatedAt: now}
	hepatocytes := &category.Category{ID: 9, Name: "Hepatocytes", Type: category.CategoryTissue, ParentID: &liverID, Description: "Liver parenchymal cells", CreatedAt: now, UpdatedAt: now}
	neurons := &category.Category{ID: 10, Name: "Neurons", Type: category.CategoryTissue, ParentID: &brainID, Description: "Nerve cells", CreatedAt: now, UpdatedAt: now}
	
	// Staining categories
	he := &category.Category{ID: 11, Name: "H&E", Type: category.CategoryStain, Description: "Hematoxylin and Eosin", CreatedAt: now, UpdatedAt: now}
	masson := &category.Category{ID: 12, Name: "Masson's Trichrome", Type: category.CategoryStain, Description: "Collagen staining", CreatedAt: now, UpdatedAt: now}
	pas := &category.Category{ID: 13, Name: "PAS", Type: category.CategoryStain, Description: "Periodic acid-Schiff", CreatedAt: now, UpdatedAt: now}
	
	r.categories = map[uint]*category.Category{
		1: human, 2: mouse, 3: rat,
		4: heart, 5: liver, 6: brain, 7: kidney,
		8: cardiacMuscle, 9: hepatocytes, 10: neurons,
		11: he, 12: masson, 13: pas,
	}
	r.nextID = 14
}