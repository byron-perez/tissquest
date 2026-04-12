package repositories

import (
	"os"
	"strings"

	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/core/taxon"
	"mcba/tissquest/internal/core/tissuerecord"
)

func NewAtlasRepository() atlas.RepositoryInterface {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "postgres" || dbType == "postgresql" {
		return NewPostgresAtlasRepository()
	}
	return NewGormAtlasRepository()
}

func NewTissueRecordRepository() tissuerecord.RepositoryInterface {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "postgres" || dbType == "postgresql" {
		return NewPostgresTissueRecordRepository()
	}
	return NewGormTissueRecordRepository()
}

func NewTaxonRepo() taxon.RepositoryInterface {
	return NewTaxonRepository()
}

func NewCategoryRepository() category.RepositoryInterface {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "postgres" || dbType == "postgresql" {
		return NewPostgresCategoryRepository()
	}
	return NewGormCategoryRepository()
}
