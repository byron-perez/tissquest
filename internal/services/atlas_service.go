package services

import (
	"mcba/tissquest/internal/core/atlas"
	"time"
)

type AtlasService struct {
	repo atlas.RepositoryInterface
}

func NewAtlasService(repo atlas.RepositoryInterface) *AtlasService {
	return &AtlasService{repo: repo}
}

func (s *AtlasService) CreateAtlas(a *atlas.Atlas) (uint, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	return s.repo.Save(a), nil
}

func (s *AtlasService) GetAtlas(id uint) (*atlas.Atlas, error) {
	return s.repo.Retrieve(id)
}

func (s *AtlasService) UpdateAtlas(id uint, a *atlas.Atlas) error {
	if err := a.Validate(); err != nil {
		return err
	}
	a.UpdatedAt = time.Now()
	return s.repo.Update(id, a)
}

func (s *AtlasService) DeleteAtlas(id uint) error {
	return s.repo.Delete(id)
}

func (s *AtlasService) ListAtlases() ([]atlas.Atlas, error) {
	return s.repo.List()
}

func (s *AtlasService) FindAtlasByName(name string) ([]atlas.Atlas, error) {
	return s.repo.FindByName(name)
}

func (s *AtlasService) ListAtlasesWithPagination(page, pageSize int) ([]atlas.Atlas, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return s.repo.ListWithPagination(page, pageSize)
}
