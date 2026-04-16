package slide

type RepositoryInterface interface {
	Save(sl *Slide) (uint, error)
	GetByID(id uint) (*Slide, error)
	Update(id uint, sl *Slide) error
	Delete(id uint) error
	ListByTissueRecord(tissueRecordID uint) ([]Slide, error)
}
