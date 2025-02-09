package tissuerecord

import (
	"mcba/tissquest/internal/core/slide"
)

type TissueRecord struct {
	Name           string
	Notes          string
	Taxonomicclass string
	Slides         []slide.Slide
	repository     RepositoryInterface
}

func (tr *TissueRecord) ConfigureTissueRecord(repository RepositoryInterface) {
	tr.repository = repository
}

func (tr *TissueRecord) Save() uint {
	persistence_response := tr.repository.Save(tr)
	return persistence_response
}

func (tr *TissueRecord) GetById(id uint) TissueRecord {
	persistence_response := tr.repository.Retrieve(id)
	return persistence_response
}

func (tr *TissueRecord) Update(id uint, tissuerecord TissueRecord) {
	tr.repository.Update(id, &tissuerecord)
}
