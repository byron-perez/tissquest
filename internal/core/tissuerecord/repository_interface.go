package tissuerecord

import (
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/core/category"
)

type RepositoryInterface interface {
	Save(tr *TissueRecord) uint
	Retrieve(id uint) (TissueRecord, int)
	Update(id uint, tr *TissueRecord)
	Delete(id uint)
	List(page, limit int) ([]TissueRecord, int64, error)

	AddAtlas(trID, atlasID uint) error
	RemoveAtlas(trID, atlasID uint) error
	ListAtlases(trID uint) ([]atlas.Atlas, error)

	AddCategory(trID, catID uint) error
	RemoveCategory(trID, catID uint) error
	ListCategories(trID uint) ([]category.Category, error)
}
