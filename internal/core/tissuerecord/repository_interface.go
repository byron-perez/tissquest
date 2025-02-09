package tissuerecord

type RepositoryInterface interface {
	Save(tr *TissueRecord) uint
	Retrieve(id uint) TissueRecord
	Update(id uint, tr *TissueRecord)
	Delete(id uint)
}
