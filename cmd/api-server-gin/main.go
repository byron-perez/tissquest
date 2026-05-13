package main

import (
	"fmt"
	"log"
	"mcba/tissquest/cmd/api-server-gin/categories"
	"mcba/tissquest/cmd/api-server-gin/collections"
	"mcba/tissquest/cmd/api-server-gin/index"
	"mcba/tissquest/cmd/api-server-gin/slides"
	"mcba/tissquest/cmd/api-server-gin/taxa"
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

	// Collection routes
	r.GET("/collections", collections.ListCollections)
	r.GET("/collections/new", collections.NewCollectionForm)
	r.GET("/collections/new-form-cancel", collections.NewCollectionFormCancel)
	r.POST("/collections", collections.CreateCollection)
	r.GET("/collections/:id/edit", collections.EditCollectionForm)
	r.GET("/collections/:id/edit-cancel", collections.EditCancelCollection)
	r.PUT("/collections/:id", collections.UpdateCollection)
	r.DELETE("/collections/:id", collections.DeleteCollection)
	r.GET("/collections/:id/confirm-delete", collections.ConfirmDeleteCollection)
	r.GET("/collections/:id/confirm-delete-cancel", collections.ConfirmDeleteCollectionCancel)
	r.GET("/collections/:id", collections.ViewCollection)
	r.GET("/collections/:id/builder", collections.BuilderPage)
	r.PUT("/collections/:id/metadata", collections.UpdateCollectionMetadata)

	// Section routes
	r.GET("/collections/:id/sections/new", collections.NewSectionForm)
	r.GET("/collections/:id/sections/new-cancel", collections.NewSectionFormCancel)
	r.POST("/collections/:id/sections", collections.CreateSection)
	r.PUT("/collections/:id/sections/:sid", collections.UpdateSection)
	r.DELETE("/collections/:id/sections/:sid", collections.DeleteSection)
	r.POST("/collections/:id/sections/reorder", collections.ReorderSections)
	r.POST("/collections/:id/sections/:sid/move", collections.MoveSection)

	// Assignment routes
	r.POST("/collections/:id/sections/:sid/assignments", collections.CreateAssignment)
	r.DELETE("/collections/:id/sections/:sid/assignments/:aid", collections.DeleteAssignment)
	r.POST("/collections/:id/sections/:sid/assignments/reorder", collections.ReorderAssignments)
	r.POST("/collections/:id/sections/:sid/assignments/:aid/move", collections.MoveAssignment)

	// Inline TR creation with assignment
	r.POST("/collections/:id/sections/:sid/tissue_records", collections.CreateTissueRecordAndAssign)

	// Tissue record routes
	r.GET("/tissue_records", tissue_records.ExplorerPage)
	r.GET("/tissue_records/search", tissue_records.SearchTissueRecords)
	r.GET("/tissue_records/new", tissue_records.NewTissueRecordForm)
	r.GET("/tissue_records/new-form-cancel", tissue_records.NewTissueRecordFormCancel)
	r.POST("/tissue_records", tissue_records.CreateTissueRecordHTML)
	r.GET("/tissue_records/:id", tissue_records.RedirectToWorkspace)
	r.GET("/tissue_records/:id/workspace", tissue_records.WorkspaceHandler)
	r.GET("/tissue_records/:id/workspace/basic-info", tissue_records.BasicInfoFragment)
	r.PUT("/tissue_records/:id/workspace/basic-info", tissue_records.SaveBasicInfo)
	r.GET("/tissue_records/:id/collections-section", tissue_records.CollectionSectionFragment)
	r.POST("/tissue_records/:id/categories/:categoryID", tissue_records.AddCategoryToTissueRecord)
	r.DELETE("/tissue_records/:id/categories/:categoryID", tissue_records.RemoveCategoryFromTissueRecord)
	r.GET("/tissue_records/:id/categories-section", tissue_records.CategorySectionFragment)
	r.GET("/tissue_records/:id/edit", tissue_records.EditTissueRecordForm)
	r.GET("/tissue_records/:id/edit-cancel", tissue_records.EditCancelTissueRecord)
	r.PUT("/tissue_records/:id", tissue_records.UpdateTissueRecordHTML)
	r.DELETE("/tissue_records/:id", tissue_records.DeleteTissueRecordHTML)
	r.GET("/tissue_records/:id/confirm-delete", tissue_records.ConfirmDeleteTissueRecord)
	r.GET("/tissue_records/:id/confirm-delete-cancel", tissue_records.ConfirmDeleteTissueRecordCancel)
	r.GET("/tissue_records/:id/slides", slides.ListSlides)
	r.GET("/tissue_records/:id/slides/new", slides.NewSlideForm)
	r.POST("/tissue_records/:id/slides", slides.CreateSlide)

	// Slide routes
	r.GET("/slides/:id/edit", slides.EditSlideForm)
	r.PUT("/slides/:id", slides.UpdateSlide)
	r.DELETE("/slides/:id", slides.DeleteSlide)
	r.GET("/slides/:id/confirm-delete", slides.ConfirmDeleteSlide)
	r.GET("/slides/:id/confirm-delete-cancel", slides.ConfirmDeleteSlideCancel)
	r.POST("/slides/:id/image", slides.UploadSlideImage(s3))
	r.PATCH("/slides/:id/images/:size", slides.SetImageVariant)
	r.GET("/api/slides/:id/dzi", slides.GetDziMetadata)
	r.PATCH("/api/slides/:id/home-viewport", slides.SetHomeViewport)
	r.GET("/slides/:id/viewer", slides.ViewSlide)
	r.POST("/slides/:id/tile", slides.TileSlide(s3))

	// Taxa routes
	r.GET("/taxa", taxa.ListTaxa)
	r.GET("/taxa/new", taxa.NewTaxonForm)
	r.GET("/taxa/new-form-cancel", taxa.NewTaxonFormCancel)
	r.POST("/taxa", taxa.CreateTaxon)
	r.GET("/taxa/:id/edit", taxa.EditTaxonForm)
	r.GET("/taxa/:id/edit-cancel", taxa.EditCancelTaxon)
	r.PUT("/taxa/:id", taxa.UpdateTaxon)
	r.DELETE("/taxa/:id", taxa.DeleteTaxon)
	r.GET("/taxa/:id/confirm-delete", taxa.ConfirmDeleteTaxon)
	r.GET("/taxa/:id/confirm-delete-cancel", taxa.ConfirmDeleteTaxonCancel)

	// Category routes
	r.GET("/categories", categories.ListCategories)
	r.GET("/categories/new", categories.NewCategoryForm)
	r.GET("/categories/new-form-cancel", categories.NewCategoryFormCancel)
	r.POST("/categories", categories.CreateCategory)
	r.GET("/categories/:id/edit", categories.EditCategoryForm)
	r.GET("/categories/:id/edit-cancel", categories.EditCancelCategory)
	r.PUT("/categories/:id", categories.UpdateCategory)
	r.DELETE("/categories/:id", categories.DeleteCategory)
	r.GET("/categories/:id/confirm-delete", categories.ConfirmDeleteCategory)
	r.GET("/categories/:id/confirm-delete-cancel", categories.ConfirmDeleteCategoryCancel)

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
	log.Println("    GET    /")
	log.Println("    --- Collections ---")
	log.Println("    GET    /collections")
	log.Println("    GET    /collections/new")
	log.Println("    POST   /collections")
	log.Println("    GET    /collections/:id")
	log.Println("    GET    /collections/:id/builder")
	log.Println("    PUT    /collections/:id/metadata")
	log.Println("    PUT    /collections/:id")
	log.Println("    DELETE /collections/:id")
	log.Println("    POST   /collections/:id/sections")
	log.Println("    PUT    /collections/:id/sections/:sid")
	log.Println("    DELETE /collections/:id/sections/:sid")
	log.Println("    POST   /collections/:id/sections/:sid/assignments")
	log.Println("    DELETE /collections/:id/sections/:sid/assignments/:aid")
	log.Println("    --- Tissue Records ---")
	log.Println("    GET    /tissue_records")
	log.Println("    GET    /tissue_records/search")
	log.Println("    GET    /tissue_records/new")
	log.Println("    POST   /tissue_records")
	log.Println("    GET    /tissue_records/:id/workspace")
	log.Println("    --- Slides ---")
	log.Println("    GET    /slides/:id/edit")
	log.Println("    PUT    /slides/:id")
	log.Println("    DELETE /slides/:id")
	log.Println("    --- Taxa ---")
	log.Println("    GET    /taxa")
	log.Println("    --- Categories ---")
	log.Println("    GET    /categories")
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
