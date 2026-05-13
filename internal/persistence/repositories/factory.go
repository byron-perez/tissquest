package repositories

import (
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/core/collection"
	"mcba/tissquest/internal/core/metacategory"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/taxon"
	"mcba/tissquest/internal/core/tissuerecord"
)

func NewCollectionRepository() collection.RepositoryInterface {
	return NewGormCollectionRepository()
}

func NewTissueRecordRepository() tissuerecord.RepositoryInterface {
	return NewGormTissueRecordRepository()
}

func NewTaxonRepository() taxon.RepositoryInterface {
	return newGormTaxonRepository()
}

func NewCategoryRepository() category.RepositoryInterface {
	return NewGormCategoryRepository()
}

func NewMetacategoryRepository() metacategory.RepositoryInterface {
	return NewGormMetacategoryRepository()
}

func NewSlideRepository() slide.RepositoryInterface {
	return NewGormSlideRepository()
}
