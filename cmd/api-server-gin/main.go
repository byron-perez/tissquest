package main

import (
	"fmt"
	"log"
	"mcba/tissquest/cmd/api-server-gin/atlas"
	"mcba/tissquest/cmd/api-server-gin/categories"
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

	// Atlas routes
	r.GET("/atlases", atlas.ListAtlasesHTML)
	r.GET("/atlases/new", atlas.NewAtlasForm)
	r.GET("/atlases/new-form-cancel", atlas.NewAtlasFormCancel)
	r.POST("/atlases", atlas.CreateAtlasHTML)
	r.GET("/atlases/:id/edit", atlas.EditAtlasForm)
	r.GET("/atlases/:id/edit-cancel", atlas.EditCancelAtlas)
	r.PUT("/atlases/:id", atlas.UpdateAtlasHTML)
	r.DELETE("/atlases/:id", atlas.DeleteAtlasHTML)
	r.GET("/atlases/:id/confirm-delete", atlas.ConfirmDeleteAtlas)
	r.GET("/atlases/:id/confirm-delete-cancel", atlas.ConfirmDeleteAtlasCancel)
	r.GET("/atlas/:id", atlas.ViewAtlas)

	// Tissue record routes
	r.GET("/tissue_records", tissue_records.ListTissueRecordsHTML)
	r.GET("/tissue_records/new", tissue_records.NewTissueRecordForm)
	r.GET("/tissue_records/new-form-cancel", tissue_records.NewTissueRecordFormCancel)
	r.POST("/tissue_records", tissue_records.CreateTissueRecordHTML)
	r.GET("/tissue_records/:id", tissue_records.RedirectToWorkspace)
	r.GET("/tissue_records/:id/workspace", tissue_records.WorkspaceHandler)
	r.GET("/tissue_records/:id/workspace/basic-info", tissue_records.BasicInfoFragment)
	r.PUT("/tissue_records/:id/workspace/basic-info", tissue_records.SaveBasicInfo)
	r.POST("/tissue_records/:id/atlases/:atlasID", tissue_records.AddAtlasToTissueRecord)
	r.DELETE("/tissue_records/:id/atlases/:atlasID", tissue_records.RemoveAtlasFromTissueRecord)
	r.GET("/tissue_records/:id/atlases-section", tissue_records.AtlasSectionFragment)
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
	log.Println("    --- Atlases ---")
	log.Println("    GET    /atlases")
	log.Println("    GET    /atlases/new")
	log.Println("    GET    /atlases/new-form-cancel")
	log.Println("    POST   /atlases")
	log.Println("    GET    /atlases/:id/edit")
	log.Println("    PUT    /atlases/:id")
	log.Println("    DELETE /atlases/:id")
	log.Println("    GET    /atlases/:id/confirm-delete")
	log.Println("    GET    /atlases/:id/confirm-delete-cancel")
	log.Println("    GET    /atlas/:id")
	log.Println("    --- Tissue Records ---")
	log.Println("    GET    /tissue_records")
	log.Println("    GET    /tissue_records/new")
	log.Println("    GET    /tissue_records/new-form-cancel")
	log.Println("    POST   /tissue_records")
	log.Println("    GET    /tissue_records/:id")
	log.Println("    GET    /tissue_records/:id/edit")
	log.Println("    PUT    /tissue_records/:id")
	log.Println("    DELETE /tissue_records/:id")
	log.Println("    GET    /tissue_records/:id/confirm-delete")
	log.Println("    GET    /tissue_records/:id/confirm-delete-cancel")
	log.Println("    GET    /tissue_records/:id/slides")
	log.Println("    GET    /tissue_records/:id/slides/new")
	log.Println("    POST   /tissue_records/:id/slides")
	log.Println("    --- Slides ---")
	log.Println("    GET    /slides/:id/edit")
	log.Println("    PUT    /slides/:id")
	log.Println("    DELETE /slides/:id")
	log.Println("    GET    /slides/:id/confirm-delete")
	log.Println("    GET    /slides/:id/confirm-delete-cancel")
	log.Println("    POST   /slides/:id/image")
	log.Println("    --- Taxa ---")
	log.Println("    GET    /taxa")
	log.Println("    GET    /taxa/new")
	log.Println("    GET    /taxa/new-form-cancel")
	log.Println("    POST   /taxa")
	log.Println("    GET    /taxa/:id/edit")
	log.Println("    PUT    /taxa/:id")
	log.Println("    DELETE /taxa/:id")
	log.Println("    GET    /taxa/:id/confirm-delete")
	log.Println("    GET    /taxa/:id/confirm-delete-cancel")
	log.Println("    --- Categories ---")
	log.Println("    GET    /categories")
	log.Println("    GET    /categories/new")
	log.Println("    GET    /categories/new-form-cancel")
	log.Println("    POST   /categories")
	log.Println("    GET    /categories/:id/edit")
	log.Println("    PUT    /categories/:id")
	log.Println("    DELETE /categories/:id")
	log.Println("    GET    /categories/:id/confirm-delete")
	log.Println("    GET    /categories/:id/confirm-delete-cancel")
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
