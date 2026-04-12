package tissuerecord

import (
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/taxon"
)

type TissueRecord struct {
	ID     uint
	Name   string
	Notes  string
	TaxonID *uint
	Taxon  *taxon.Taxon
	Slides []slide.Slide
}
