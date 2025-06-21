package tissuerecord

type RepositoryInterface interface {
	Save(tr *TissueRecord) uint
	Retrieve(id uint) (TissueRecord, int)
	Update(id uint, tr *TissueRecord)
	Delete(id uint)
	List(page, limit int) ([]TissueRecord, int64, error)
}
