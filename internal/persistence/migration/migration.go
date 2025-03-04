package migration

import (
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Tabler interface {
	TableName() string
}

func RunMigration() {
	db, err := gorm.Open(sqlite.Open(os.Getenv("DB_PATH")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&TissueRecordModel{}, &SlideModel{})
}
