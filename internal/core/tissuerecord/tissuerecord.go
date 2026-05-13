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
	// FeaturedImageURL is a read-only field populated by the repository
	// when listing records for display. It holds the best available
	// thumbnail URL across all slides of this record.
	FeaturedImageURL string
}
