package repositories

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbInstance *gorm.DB
	mu         sync.Mutex
)

// GetDB ensures a singleton connection pool that can recover from initial failures.
func GetDB() (*gorm.DB, error) {
	// Fast path: If already initialized, return immediately without locking
	if dbInstance != nil {
		return dbInstance, nil
	}

	mu.Lock()
	defer mu.Unlock()

	// Double-check to prevent race conditions if two routines locked at once
	if dbInstance != nil {
		return dbInstance, nil
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=UTC",
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_PORT"),
	)

	// Open connection with GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Speeds up repeated JOIN queries by caching statements
		Logger:      logger.Default.LogMode(logger.Silent), // Keep logs clean for production
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to Aurora: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// --- TUNING FOR 50 USERS ON AURORA FREE TIER ---
	
	// MaxOpen: 25 is the "Magic Number" for 50 users. 
	// Statistically, half your users are reading the screen while the other half hit the DB.
	sqlDB.SetMaxOpenConns(25)

	// MaxIdle: Keeps 5-10 connections open and "warm" so users don't feel a 100ms lag 
	// while a new connection handshakes with AWS.
	sqlDB.SetMaxIdleConns(10)

	// MaxLifetime: Keeps the Aurora instance memory fresh by recycling connections.
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Health Check: Ensure the connection is actually valid before giving it to the app
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	dbInstance = db
	return dbInstance, nil
}
