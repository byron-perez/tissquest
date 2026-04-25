package migration

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Tabler interface {
	TableName() string
}

var (
	migrationDB   *gorm.DB
	migrationOnce sync.Once
)

func openDB() (*gorm.DB, error) {
	var initErr error
	migrationOnce.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=UTC",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_PASSWORD"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_PORT"),
		)
		// Raise slow-query threshold so AutoMigrate's catalog introspection
		// doesn't flood logs with false positives on Aurora Free Tier.
		migrationLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             500 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: migrationLogger})
		if err != nil {
			initErr = err
			return
		}
		sqlDB, err := db.DB()
		if err != nil {
			initErr = err
			return
		}
		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		migrationDB = db
	})
	return migrationDB, initErr
}

func RunMigration() {
	db, err := openDB()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to PostgreSQL database: %v", err))
	}
	fmt.Println("Connected to PostgreSQL database")

	// Drop the old atlases table only if the collections table doesn't exist yet.
	// This ensures we only do this once (on first migration after the rename).
	if !db.Migrator().HasTable("collections") {
		if db.Migrator().HasTable("atlases") {
			if err := db.Migrator().DropTable("atlases"); err != nil {
				panic(fmt.Sprintf("failed to drop atlases table: %v", err))
			}
			fmt.Println("Dropped legacy atlases table")
		}
	}

	if err = db.AutoMigrate(
		&TaxonModel{},
		&CategoryModel{},
		&CollectionModel{},
		&CollectionSectionModel{},
		&CollectionSectionAssignmentModel{},
		&TissueRecordModel{},
		&PreparationModel{},
		&SlideModel{},
		&SlideImageVariantModel{},
	); err != nil {
		panic(fmt.Sprintf("database migration failed: %v", err))
	}

	if err := seedDefaultCategories(db); err != nil {
		panic(fmt.Sprintf("failed to seed categories: %v", err))
	}

	if err := seedTaxa(db); err != nil {
		panic(fmt.Sprintf("failed to seed taxa: %v", err))
	}

	if err := seedSampleTissueRecords(db); err != nil {
		panic(fmt.Sprintf("failed to seed tissue records: %v", err))
	}

	fmt.Println("Database migration completed successfully")
}

func seedTaxa(db *gorm.DB) error {
	var count int64
	if err := db.Model(&TaxonModel{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	plantae := TaxonModel{Rank: "kingdom", Name: "Plantae"}
	if err := db.Create(&plantae).Error; err != nil {
		return err
	}
	tracheophyta := TaxonModel{Rank: "phylum", Name: "Tracheophyta", ParentID: &plantae.ID}
	if err := db.Create(&tracheophyta).Error; err != nil {
		return err
	}
	polypodiopsida := TaxonModel{Rank: "class", Name: "Polypodiopsida", ParentID: &tracheophyta.ID}
	magnoliopsida := TaxonModel{Rank: "class", Name: "Magnoliopsida", ParentID: &tracheophyta.ID}
	return db.Create(&[]TaxonModel{polypodiopsida, magnoliopsida}).Error
}

func seedSampleTissueRecords(db *gorm.DB) error {
	var count int64
	if err := db.Model(&TissueRecordModel{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var hoja, parenquima, xe, he, azul CategoryModel

	if err := db.Where("name = ?", "Hoja").First(&hoja).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "Par\u00e9nquima").First(&parenquima).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "Xilema").First(&xe).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "H&E").First(&he).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "Azul de metileno").First(&azul).Error; err != nil {
		return err
	}

	var polypodiopsida, magnoliopsida TaxonModel
	if err := db.Where("name = ? AND rank = ?", "Polypodiopsida", "class").First(&polypodiopsida).Error; err != nil {
		return err
	}
	if err := db.Where("name = ? AND rank = ?", "Magnoliopsida", "class").First(&magnoliopsida).Error; err != nil {
		return err
	}

	fernRecord := TissueRecordModel{
		Name:    "Fronda de helecho",
		Notes:   "Corte longitudinal y transversal de un helecho (Pteridium sp.), preparado para mostrar la anatom\u00eda de la fronda y los tejidos internos.",
		TaxonID: &polypodiopsida.ID,
		Slides: []SlideModel{
			{
				Name:          "Corte longitudinal",
				Magnification: 40,
				Preparation:   PreparationModel{Staining: "H&E", InclusionMethod: "Parafina", Reagents: "Hematoxilina, Eosina", Protocol: "Deshidrataci\u00f3n en etanol, inclusi\u00f3n en parafina, secci\u00f3n 5\u03bcm"},
			},
			{
				Name:          "Corte transversal",
				Magnification: 100,
				Preparation:   PreparationModel{Staining: "Azul de metileno", InclusionMethod: "Criost\u00e1to", Reagents: "Azul de metileno 1%", Protocol: "Secci\u00f3n en fresco, tinci\u00f3n directa"},
			},
		},
	}

	stemRecord := TissueRecordModel{
		Name:    "Corte de tallo",
		Notes:   "Secci\u00f3n transversal de tallo vascular mostrando xilema y floema, \u00fatil para entender conducci\u00f3n y organizaci\u00f3n de tejidos.",
		TaxonID: &magnoliopsida.ID,
		Slides: []SlideModel{
			{
				Name:          "Tallo transversal",
				Magnification: 80,
				Preparation:   PreparationModel{Staining: "PAS", InclusionMethod: "Parafina", Reagents: "\u00c1cido peri\u00f3dico, reactivo de Schiff", Protocol: "Oxidaci\u00f3n con \u00e1cido peri\u00f3dico, tinci\u00f3n con Schiff"},
			},
		},
	}

	if err := db.Create(&fernRecord).Error; err != nil {
		return err
	}
	if err := db.Create(&stemRecord).Error; err != nil {
		return err
	}

	if err := db.Model(&fernRecord).Association("Categories").Append(&hoja, &parenquima, &he); err != nil {
		return err
	}
	if err := db.Model(&stemRecord).Association("Categories").Append(&xe, &azul); err != nil {
		return err
	}

	// Seed a sample collection
	collectionModel := CollectionModel{
		Name:        "Atlas b\u00e1sico de anatom\u00eda vegetal",
		Description: "Colecci\u00f3n introductoria de registros de tejido y tinciones para estudiar anatom\u00eda vegetal.",
		Type:        "atlas",
	}
	if err := db.Create(&collectionModel).Error; err != nil {
		return err
	}

	// Create a section and assign tissue records
	section := CollectionSectionModel{
		CollectionID: collectionModel.ID,
		Name:         "Tejidos vegetales",
		Position:     1,
	}
	if err := db.Create(&section).Error; err != nil {
		return err
	}

	assign1 := CollectionSectionAssignmentModel{
		SectionID:      section.ID,
		TissueRecordID: fernRecord.ID,
		Position:       1,
	}
	assign2 := CollectionSectionAssignmentModel{
		SectionID:      section.ID,
		TissueRecordID: stemRecord.ID,
		Position:       2,
	}
	if err := db.Create(&assign1).Error; err != nil {
		return err
	}
	if err := db.Create(&assign2).Error; err != nil {
		return err
	}

	return nil
}

func seedDefaultCategories(db *gorm.DB) error {
	var count int64
	if err := db.Model(&CategoryModel{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	organRoot := CategoryModel{Name: "\u00d3rganos", Type: "organ", Description: "Clasificaci\u00f3n por \u00f3rgano"}
	tissueRoot := CategoryModel{Name: "Tejidos", Type: "tissue", Description: "Tipos de tejido vegetal"}
	stainRoot := CategoryModel{Name: "Tinciones", Type: "stain", Description: "T\u00e9cnicas de tinci\u00f3n"}
	taxonomyRoot := CategoryModel{Name: "Taxonom\u00eda", Type: "species", Description: "Clasificaci\u00f3n taxon\u00f3mica"}

	if err := db.Create(&[]CategoryModel{organRoot, tissueRoot, stainRoot, taxonomyRoot}).Error; err != nil {
		return err
	}

	if err := db.Where("name = ?", "\u00d3rganos").First(&organRoot).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "Tejidos").First(&tissueRoot).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "Tinciones").First(&stainRoot).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "Taxonom\u00eda").First(&taxonomyRoot).Error; err != nil {
		return err
	}

	children := []CategoryModel{
		{Name: "Ra\u00edz", Type: "organ", Description: "\u00d3rgano radicular", ParentID: &organRoot.ID},
		{Name: "Tallo", Type: "organ", Description: "\u00d3rgano caulinar", ParentID: &organRoot.ID},
		{Name: "Hoja", Type: "organ", Description: "\u00d3rgano foliar", ParentID: &organRoot.ID},
		{Name: "Xilema", Type: "tissue", Description: "Tejido conductor de agua", ParentID: &tissueRoot.ID},
		{Name: "Floema", Type: "tissue", Description: "Tejido conductor de nutrientes", ParentID: &tissueRoot.ID},
		{Name: "Par\u00e9nquima", Type: "tissue", Description: "Tejido de almacenamiento y soporte", ParentID: &tissueRoot.ID},
		{Name: "H&E", Type: "stain", Description: "Hematoxilina y eosina", ParentID: &stainRoot.ID},
		{Name: "PAS", Type: "stain", Description: "Periodic Acid-Schiff", ParentID: &stainRoot.ID},
		{Name: "Azul de metileno", Type: "stain", Description: "Tinci\u00f3n de metileno azul", ParentID: &stainRoot.ID},
		{Name: "Plantae", Type: "species", Description: "Reino de las plantas", ParentID: &taxonomyRoot.ID},
		{Name: "Magnoliophyta", Type: "species", Description: "Plantas con flor", ParentID: &taxonomyRoot.ID},
	}

	return db.Create(&children).Error
}
