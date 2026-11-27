package db

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
	mu       sync.Mutex
	instance *gorm.DB
)

func DSN() string {
	sslmode := os.Getenv("DATABASE_SSLMODE")
	if sslmode == "" {
		sslmode = "require"
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC connect_timeout=10",
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_PORT"),
		sslmode,
	)
}

func Get() (*gorm.DB, error) {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return instance, nil
	}

	gdb, err := open()
	if err != nil {
		return nil, err
	}

	instance = gdb
	return instance, nil
}

func open() (*gorm.DB, error) {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		gdb, err := gorm.Open(postgres.Open(DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		sqlDB, err := gdb.DB()
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)

		if err := sqlDB.Ping(); err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		return gdb, nil
	}

	return nil, fmt.Errorf("database connection failed after retries: %w", lastErr)
}
