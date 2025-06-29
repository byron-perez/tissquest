package migration

import (
    "fmt"
    "os"
    "strings"

    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type Tabler interface {
    TableName() string
}

// RunMigration sets up the database connection and runs migrations
// based on the DB_TYPE environment variable
func RunMigration() {
    dbType := strings.ToLower(os.Getenv("DB_TYPE"))
    if dbType == "" {
        dbType = "sqlite" // Default to SQLite if not specified
    }

    var db *gorm.DB
    var err error

    switch dbType {
    case "postgres", "postgresql":
        dsn := fmt.Sprintf(
            "host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=UTC",
            os.Getenv("DATABASE_HOST"),
            os.Getenv("DATABASE_USER"),
            os.Getenv("DATABASE_PASSWORD"),
            os.Getenv("DATABASE_NAME"),
            os.Getenv("DATABASE_PORT"),
        )
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
        if err != nil {
            panic(fmt.Sprintf("failed to connect to PostgreSQL database: %v", err))
        }
        fmt.Println("Connected to PostgreSQL database")

    case "sqlite":
        dbPath := os.Getenv("DB_PATH")
        if dbPath == "" {
            dbPath = "tissquest.db" // Default SQLite database path
        }
        db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
        if err != nil {
            panic(fmt.Sprintf("failed to connect to SQLite database: %v", err))
        }
        fmt.Println("Connected to SQLite database")

    default:
        panic(fmt.Sprintf("unsupported database type: %s", dbType))
    }

    // Run migrations for all models
    err = db.AutoMigrate(
        &TissueRecordModel{},
        &SlideModel{},
        // Add any other models that need migration here
    )
    if err != nil {
        panic(fmt.Sprintf("database migration failed: %v", err))
    }
    fmt.Println("Database migration completed successfully")
}