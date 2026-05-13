package services

import (
	"mcba/tissquest/internal/core/metacategory"
	"time"
)

// MetacategoryHierarchy represents a metacategory with its child items at the current level
type MetacategoryHierarchy struct {
	ID          uint
	Name        string
	Description string
	Children    []MetacategoryNode
}

// MetacategoryNode represents a node in the hierarchy
type MetacategoryNode struct {
	ID          uint
	Name        string
	Description string
	ParentID    *uint
	HasChildren bool
}

type MetacategoryService struct {
	repo metacategory.RepositoryInterface
}

func NewMetacategoryService(repo metacategory.RepositoryInterface) *MetacategoryService {
	return &MetacategoryService{repo: repo}
}

func (s *MetacategoryService) Create(m *metacategory.Metacategory) (uint, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	return s.repo.Save(m)
}

func (s *MetacategoryService) GetByID(id uint) (*metacategory.Metacategory, error) {
	return s.repo.Retrieve(id)
}

func (s *MetacategoryService) Update(id uint, m *metacategory.Metacategory) error {
	if err := m.Validate(); err != nil {
		return err
	}
	m.UpdatedAt = time.Now()
	return s.repo.Update(id, m)
}

func (s *MetacategoryService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *MetacategoryService) List() ([]metacategory.Metacategory, error) {
	return s.repo.List()
}

func (s *MetacategoryService) GetRootMetacategories() ([]metacategory.Metacategory, error) {
	return s.repo.FindRootMetacategories()
}

func (s *MetacategoryService) GetChildren(parentID uint) ([]metacategory.Metacategory, error) {
	return s.repo.FindByParent(parentID)
}

func (s *MetacategoryService) GetFullHierarchy(id uint) (*metacategory.Metacategory, error) {
	return s.repo.GetFullHierarchy(id)
}

func (s *MetacategoryService) ListWithChildren() ([]metacategory.Metacategory, error) {
	return s.repo.ListWithChildren()
}

// GetHierarchiesForDisplay returns a structured view of root metacategories with their immediate children
func (s *MetacategoryService) GetHierarchiesForDisplay() ([]MetacategoryHierarchy, error) {
	roots, err := s.repo.ListWithChildren()
	if err != nil {
		return nil, err
	}

	hierarchies := make([]MetacategoryHierarchy, 0, len(roots))
	for _, root := range roots {
		h := MetacategoryHierarchy{
			ID:          root.ID,
			Name:        root.Name,
			Description: root.Description,
			Children:    make([]MetacategoryNode, 0, len(root.Children)),
		}

		for _, child := range root.Children {
			h.Children = append(h.Children, MetacategoryNode{
				ID:          child.ID,
				Name:        child.Name,
				Description: child.Description,
				ParentID:    child.ParentID,
				HasChildren: child.HasChildren(),
			})
		}

		hierarchies = append(hierarchies, h)
	}

	return hierarchies, nil
}
