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

func RunMigration() {
    dbType := strings.ToLower(os.Getenv("DB_TYPE"))
    if dbType == "" {
        dbType = "sqlite"
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
            dbPath = "tissquest.db"
        }
        db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
        if err != nil {
            panic(fmt.Sprintf("failed to connect to SQLite database: %v", err))
        }
        fmt.Println("Connected to SQLite database")

    default:
        panic(fmt.Sprintf("unsupported database type: %s", dbType))
    }

    if err = db.AutoMigrate(
        &TaxonModel{},
        &CategoryModel{},
        &AtlasModel{},
        &TissueRecordModel{},
        &PreparationModel{},
        &SlideModel{},
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
                Url:           "https://botweb.uwsp.edu/Anatomy/images/dicotwood/images_c/Anat0343.jpg",
                Magnification: 40,
                Preparation:   PreparationModel{Staining: "H&E", InclusionMethod: "Parafina", Reagents: "Hematoxilina, Eosina", Protocol: "Deshidrataci\u00f3n en etanol, inclusi\u00f3n en parafina, secci\u00f3n 5\u03bcm"},
            },
            {
                Name:          "Corte transversal",
                Url:           "https://botweb.uwsp.edu/Anatomy/images/primaryxylem/images_c/Anat0144.jpg",
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
                Url:           "https://upload.wikimedia.org/wikipedia/commons/5/5d/Stem_cross_section.png",
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

    atlasModel := AtlasModel{
        Name:        "Atlas b\u00e1sico de anatom\u00eda vegetal",
        Description: "Colecci\u00f3n introductoria de registros de tejido y tinciones para estudiar anatom\u00eda vegetal.",
        Category:    "Bot\u00e1nica",
    }
    if err := db.Create(&atlasModel).Error; err != nil {
        return err
    }

    if err := db.Model(&atlasModel).Association("TissueRecords").Append(&fernRecord, &stemRecord); err != nil {
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
