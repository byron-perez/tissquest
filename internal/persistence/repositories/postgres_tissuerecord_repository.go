package repositories

// PostgresTissueRecordRepository is an alias for GormTissueRecordRepository.
// All tissue record persistence is handled by GormTissueRecordRepository
// which resolves the correct driver from DB_TYPE env var.
type PostgresTissueRecordRepository = GormTissueRecordRepository

func NewPostgresTissueRecordRepository() *GormTissueRecordRepository {
	return NewGormTissueRecordRepository()
}
