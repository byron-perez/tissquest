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
        &CategoryModel{},
        &AtlasModel{},
        &TissueRecordModel{},
        &SlideModel{},
        &StainingModel{},
    ); err != nil {
        panic(fmt.Sprintf("database migration failed: %v", err))
    }

    if err := seedDefaultCategories(db); err != nil {
        panic(fmt.Sprintf("failed to seed categories: %v", err))
    }

    if err := seedSampleTissueRecords(db); err != nil {
        panic(fmt.Sprintf("failed to seed tissue records: %v", err))
    }

    fmt.Println("Database migration completed successfully")
}

func seedSampleTissueRecords(db *gorm.DB) error {
    var count int64
    if err := db.Model(&TissueRecordModel{}).Count(&count).Error; err != nil {
        return err
    }
    if count > 0 {
        return nil
    }

    var hoja CategoryModel
    var parenquima CategoryModel
    var xe CategoryModel
    var he CategoryModel
    var azul CategoryModel
    var plantae CategoryModel
    var magnoliophyta CategoryModel

    if err := db.Where("name = ?", "Hoja").First(&hoja).Error; err != nil {
        return err
    }
    if err := db.Where("name = ?", "Parénquima").First(&parenquima).Error; err != nil {
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
    if err := db.Where("name = ?", "Plantae").First(&plantae).Error; err != nil {
        return err
    }
    if err := db.Where("name = ?", "Magnoliophyta").First(&magnoliophyta).Error; err != nil {
        return err
    }

    fernRecord := TissueRecordModel{
        Name:           "Fronda de helecho",
        Notes:          "Corte longitudinal y transversal de un helecho (Pteridium sp.), preparado para mostrar la anatomía de la fronda y los tejidos internos.",
        Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
        Slides: []SlideModel{
            {Name: "Corte longitudinal", Url: "https://botweb.uwsp.edu/Anatomy/images/dicotwood/images_c/Anat0343.jpg", Staining: StainingModel{Name: "H&E"}, Magnification: 40},
            {Name: "Corte transversal", Url: "https://botweb.uwsp.edu/Anatomy/images/primaryxylem/images_c/Anat0144.jpg", Staining: StainingModel{Name: "Azul de metileno"}, Magnification: 100},
        },
    }

    stemRecord := TissueRecordModel{
        Name:           "Corte de tallo",
        Notes:          "Sección transversal de tallo vascular mostrando xilema y floema, útil para entender conducción y organización de tejidos.",
        Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Magnoliophyta,Cls:Magnoliopsida",
        Slides: []SlideModel{
            {Name: "Tallo transversal", Url: "https://upload.wikimedia.org/wikipedia/commons/5/5d/Stem_cross_section.png", Staining: StainingModel{Name: "PAS"}, Magnification: 80},
        },
    }

    if err := db.Create(&fernRecord).Error; err != nil {
        return err
    }
    if err := db.Create(&stemRecord).Error; err != nil {
        return err
    }

    if err := db.Model(&fernRecord).Association("Categories").Append(&hoja, &parenquima, &he, &plantae); err != nil {
        return err
    }
    if err := db.Model(&stemRecord).Association("Categories").Append(&xe, &plantae, &magnoliophyta, &azul); err != nil {
        return err
    }

    atlasModel := AtlasModel{
        Name:        "Atlas básico de anatomía vegetal",
        Description: "Colección introductoria de registros de tejido y tinciones para estudiar anatomía vegetal.",
        Category:    "Botánica",
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

    organRoot := CategoryModel{Name: "Órganos", Type: "organ", Description: "Clasificación por órgano"}
    tissueRoot := CategoryModel{Name: "Tejidos", Type: "tissue", Description: "Tipos de tejido vegetal"}
    stainRoot := CategoryModel{Name: "Tinciones", Type: "stain", Description: "Técnicas de tinción"}
    taxonomyRoot := CategoryModel{Name: "Taxonomía", Type: "species", Description: "Clasificación taxonómica"}

    if err := db.Create(&[]CategoryModel{organRoot, tissueRoot, stainRoot, taxonomyRoot}).Error; err != nil {
        return err
    }

    if err := db.Where("name = ?", "Órganos").First(&organRoot).Error; err != nil {
        return err
    }
    if err := db.Where("name = ?", "Tejidos").First(&tissueRoot).Error; err != nil {
        return err
    }
    if err := db.Where("name = ?", "Tinciones").First(&stainRoot).Error; err != nil {
        return err
    }
    if err := db.Where("name = ?", "Taxonomía").First(&taxonomyRoot).Error; err != nil {
        return err
    }

    children := []CategoryModel{
        {Name: "Raíz", Type: "organ", Description: "Órgano radicular", ParentID: &organRoot.ID},
        {Name: "Tallo", Type: "organ", Description: "Órgano caulinar", ParentID: &organRoot.ID},
        {Name: "Hoja", Type: "organ", Description: "Órgano foliar", ParentID: &organRoot.ID},
        {Name: "Xilema", Type: "tissue", Description: "Tejido conductor de agua", ParentID: &tissueRoot.ID},
        {Name: "Floema", Type: "tissue", Description: "Tejido conductor de nutrientes", ParentID: &tissueRoot.ID},
        {Name: "Parénquima", Type: "tissue", Description: "Tejido de almacenamiento y soporte", ParentID: &tissueRoot.ID},
        {Name: "H&E", Type: "stain", Description: "Hematoxilina y eosina", ParentID: &stainRoot.ID},
        {Name: "PAS", Type: "stain", Description: "Periodic Acid-Schiff", ParentID: &stainRoot.ID},
        {Name: "Azul de metileno", Type: "stain", Description: "Tinción de metileno azul", ParentID: &stainRoot.ID},
        {Name: "Plantae", Type: "species", Description: "Reino de las plantas", ParentID: &taxonomyRoot.ID},
        {Name: "Magnoliophyta", Type: "species", Description: "Plantas con flor", ParentID: &taxonomyRoot.ID},
    }

    return db.Create(&children).Error
}