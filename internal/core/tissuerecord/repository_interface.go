package tissuerecord

import (
	"mcba/tissquest/internal/core/category"
)

type RepositoryInterface interface {
	Save(tr *TissueRecord) uint
	Retrieve(id uint) (TissueRecord, int)
	Update(id uint, tr *TissueRecord)
	Delete(id uint)
	List(page, limit int) ([]TissueRecord, int64, error)

	AddCategory(trID, catID uint) error
	RemoveCategory(trID, catID uint) error
	ListCategories(trID uint) ([]category.Category, error)
}
