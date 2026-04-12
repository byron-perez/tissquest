package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := openDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	tables, err := db.Migrator().GetTables()
	if err != nil {
		log.Fatalf("failed to list tables: %v", err)
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			log.Fatalf("failed to drop table %s: %v", table, err)
		}
		fmt.Printf("dropped: %s\n", table)
	}

	fmt.Println("All tables dropped successfully.")
}

func openDB() (*gorm.DB, error) {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "postgres" || dbType == "postgresql" {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=UTC",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_PASSWORD"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_PORT"),
		)
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "tissquest.db"
	}
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
}
