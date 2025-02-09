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

func (tr *TissueRecord) Save() bool {
	persistence_response := tr.repository.Save(tr)
	return persistence_response
}
