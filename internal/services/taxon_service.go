package services

import (
	"mcba/tissquest/internal/core/taxon"
)

type TaxonService struct {
	repo taxon.RepositoryInterface
}

func NewTaxonService(repo taxon.RepositoryInterface) *TaxonService {
	return &TaxonService{repo: repo}
}

func (s *TaxonService) Create(t *taxon.Taxon) (uint, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	return s.repo.Save(t)
}

func (s *TaxonService) GetByID(id uint) (*taxon.Taxon, error) {
	return s.repo.GetByID(id)
}

func (s *TaxonService) Update(id uint, t *taxon.Taxon) error {
	if err := t.Validate(); err != nil {
		return err
	}
	return s.repo.Update(id, t)
}

func (s *TaxonService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *TaxonService) List() ([]taxon.Taxon, error) {
	return s.repo.List()
}

func (s *TaxonService) ListByRank(rank taxon.Rank) ([]taxon.Taxon, error) {
	return s.repo.ListByRank(rank)
}
