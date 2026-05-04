package tests

import (
	"mcba/tissquest/internal/core/collection"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

// newCollectionSvc creates a CollectionService wired to the given repos.
func newCollectionSvc(
	colRepo collection.RepositoryInterface,
	trRepo tissuerecord.RepositoryInterface,
) *services.CollectionService {
	return services.NewCollectionService(colRepo, trRepo)
}

// tissuerecord alias so test file can reference the type directly.
var _ tissuerecord.RepositoryInterface = (*repositories.GormTissueRecordRepository)(nil)
