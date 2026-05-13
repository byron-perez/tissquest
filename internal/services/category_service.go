package services

import (
	"mcba/tissquest/internal/core/category"
	"time"
)

type CategoryService struct {
	repo category.RepositoryInterface
}

func NewCategoryService(repo category.RepositoryInterface) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(cat *category.Category) (uint, error) {
	if err := cat.Validate(); err != nil {
		return 0, err
	}
	now := time.Now()
	cat.CreatedAt = now
	cat.UpdatedAt = now
	id := s.repo.Save(cat)
	return id, nil
}

func (s *CategoryService) GetByID(id uint) (*category.Category, error) {
	return s.repo.Retrieve(id)
}

func (s *CategoryService) Update(id uint, cat *category.Category) error {
	if err := cat.Validate(); err != nil {
		return err
	}
	cat.UpdatedAt = time.Now()
	return s.repo.Update(id, cat)
}

func (s *CategoryService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *CategoryService) List() ([]category.Category, error) {
	return s.repo.List()
}

func (s *CategoryService) ListWithCounts() ([]category.CategoryWithCount, error) {
	return s.repo.ListWithCounts()
}
