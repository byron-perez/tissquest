package tissuerecord

import (
	"mcba/tissquest/internal/core/slide"
)

type TissueRecord struct {
	Name           string
	Notes          string
	Taxonomicclass string
	Slides         []slide.Slide
}
