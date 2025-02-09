package tissuerecord

type RepositoryInterface interface {
	Save(tr *TissueRecord) bool
	Retrieve(id string) TissueRecord
	Update(id string, tr *TissueRecord)
	Delete(id string)
}
