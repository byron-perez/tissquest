package services

import (
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/core/tissuerecord"
)

type TissueRecordService struct {
	repo tissuerecord.RepositoryInterface
}

func NewTissueRecordService(repo tissuerecord.RepositoryInterface) *TissueRecordService {
	return &TissueRecordService{repo: repo}
}

func (s *TissueRecordService) Create(tr *tissuerecord.TissueRecord) uint {
	return s.repo.Save(tr)
}

func (s *TissueRecordService) GetByID(id uint) (tissuerecord.TissueRecord, int) {
	return s.repo.Retrieve(id)
}

func (s *TissueRecordService) Update(id uint, tr *tissuerecord.TissueRecord) {
	s.repo.Update(id, tr)
}

func (s *TissueRecordService) Delete(id uint) {
	s.repo.Delete(id)
}

func (s *TissueRecordService) List(page, limit int) ([]tissuerecord.TissueRecord, int64, error) {
	return s.repo.List(page, limit)
}

func (s *TissueRecordService) AddCategory(trID, catID uint) error {
	return s.repo.AddCategory(trID, catID)
}

func (s *TissueRecordService) RemoveCategory(trID, catID uint) error {
	return s.repo.RemoveCategory(trID, catID)
}

func (s *TissueRecordService) ListCategories(trID uint) ([]category.Category, error) {
	return s.repo.ListCategories(trID)
}
