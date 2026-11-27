package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	persistencedb "mcba/tissquest/internal/persistence/db"
)

func main() {
	_ = godotenv.Load()

	db, err := persistencedb.Get()
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
