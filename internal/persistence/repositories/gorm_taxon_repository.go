package repositories

import (
	"fmt"
	"mcba/tissquest/internal/core/taxon"
	"mcba/tissquest/internal/persistence/migration"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GormTaxonRepository struct{}

func NewTaxonRepository() *GormTaxonRepository {
	return &GormTaxonRepository{}
}

func (r *GormTaxonRepository) getDB() (*gorm.DB, error) {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "postgres" || dbType == "postgresql" {
		return gorm.Open(postgres.Open(buildDSN()), &gorm.Config{})
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "tissquest.db"
	}
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
}

func (r *GormTaxonRepository) Save(t *taxon.Taxon) (uint, error) {
	db, err := r.getDB()
	if err != nil {
		return 0, err
	}
	m := migration.TaxonModel{Rank: string(t.Rank), Name: t.Name, ParentID: t.ParentID}
	if err := db.Create(&m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (r *GormTaxonRepository) GetByID(id uint) (*taxon.Taxon, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, err
	}
	var m migration.TaxonModel
	if err := db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return modelToTaxon(m), nil
}

func (r *GormTaxonRepository) GetLineage(id uint) ([]taxon.Taxon, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, err
	}
	m, err := loadWithParents(db, id)
	if err != nil {
		return nil, err
	}
	t := modelToTaxonDeep(*m)
	return t.Lineage(), nil
}

func (r *GormTaxonRepository) ListByRank(rank taxon.Rank) ([]taxon.Taxon, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, err
	}
	var models []migration.TaxonModel
	if err := db.Where("rank = ?", string(rank)).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]taxon.Taxon, len(models))
	for i, m := range models {
		result[i] = *modelToTaxon(m)
	}
	return result, nil
}

func loadWithParents(db *gorm.DB, id uint) (*migration.TaxonModel, error) {
	var m migration.TaxonModel
	if err := db.First(&m, id).Error; err != nil {
		return nil, fmt.Errorf("taxon %d not found: %w", id, err)
	}
	if m.ParentID != nil {
		parent, err := loadWithParents(db, *m.ParentID)
		if err != nil {
			return nil, err
		}
		m.Parent = parent
	}
	return &m, nil
}

func modelToTaxon(m migration.TaxonModel) *taxon.Taxon {
	return &taxon.Taxon{
		ID:       m.ID,
		Rank:     taxon.Rank(m.Rank),
		Name:     m.Name,
		ParentID: m.ParentID,
	}
}

func modelToTaxonDeep(m migration.TaxonModel) *taxon.Taxon {
	t := modelToTaxon(m)
	if m.Parent != nil {
		t.Parent = modelToTaxonDeep(*m.Parent)
	}
	return t
}
