package repositories

import (
	persistencedb "mcba/tissquest/internal/persistence/db"

	"gorm.io/gorm"
)

// GetDB returns the shared PostgreSQL connection pool.
func GetDB() (*gorm.DB, error) {
	return persistencedb.Get()
}
