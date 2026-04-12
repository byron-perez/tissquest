package main

import (
	"fmt"
	"log"
	"mcba/tissquest/cmd/api-server-gin/atlas"
	"mcba/tissquest/cmd/api-server-gin/index"
	"mcba/tissquest/cmd/api-server-gin/slides"
	"mcba/tissquest/cmd/api-server-gin/tissue_records"
	"mcba/tissquest/internal/persistence/migration"
	persistencestorage "mcba/tissquest/internal/persistence/storage"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter(s3 *persistencestorage.S3Storage) (*gin.Engine, error) {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./web/static")

	// Setup routes
	r.GET("/", index.GetIndex)
	r.GET("/tissue_records", tissue_records.ListTissueRecords)
	r.POST("/tissue_records", tissue_records.CreateTissueRecord)
	r.GET("/atlases", atlas.ListAtlases)
	r.GET("/atlas/:id", atlas.ViewAtlas)
	r.POST("/slides/:id/image", slides.UploadSlideImage(s3))

	return r, nil
}

const port = ":8080"

func logStartupInfo() {
	cwd, _ := os.Getwd()

	dbType := os.Getenv("DB_TYPE")
	dbInfo := os.Getenv("DB_PATH")
	if dbType == "postgres" {
		dbInfo = fmt.Sprintf("%s@%s:%s/%s",
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_NAME"),
		)
	}

	log.Println("---------------------------------------")
	log.Println("  TissQuest API Server")
	log.Println("---------------------------------------")
	log.Printf("  Port     : %s", port)
	log.Printf("  Mode     : %s", gin.Mode())
	log.Printf("  DB type  : %s", dbType)
	log.Printf("  DB       : %s", dbInfo)
	log.Printf("  Workdir  : %s", cwd)
	log.Println("  Routes   :")
	log.Println("    GET  /")
	log.Println("    GET  /tissue_records")
	log.Println("    POST /tissue_records")
	log.Println("    GET  /atlases")
	log.Println("    GET  /atlas/:id")
	log.Println("    POST /slides/:id/image")
	log.Println("---------------------------------------")
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	migration.RunMigration()

	s3, err := persistencestorage.NewS3Storage(
		os.Getenv("AWS_REGION"),
		os.Getenv("S3_BUCKET"),
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}

	r, err := setupRouter(s3)
	if err != nil {
		log.Fatalf("Failed to set up router: %v", err)
	}

	logStartupInfo()
	r.Run(port)
}
