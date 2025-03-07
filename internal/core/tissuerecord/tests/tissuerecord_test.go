package tests

import (
	"log"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/migration"
	"mcba/tissquest/internal/persistence/repositories"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	// load .env
	err := godotenv.Load("../../../../.env")
	if err != nil {
		log.Fatal(err.Error())
	}
	// run migrations
	migration.RunMigration()

	// run tests
	m.Run()
}

func TestTissueRecordSave(t *testing.T) {
	tissslide1 := slide.Slide{Name: "Corte longitudinal", Magnification: 40}
	tissslide1.Img.Url = "https://botweb.uwsp.edu/Anatomy/images/dicotwood/images_c/Anat0343.jpg"
	tissslide1.Staining.Name = "HyE"
	tissslide2 := slide.Slide{Name: "Corte transversal", Magnification: 100}
	tissslide2.Img.Url = "https://botweb.uwsp.edu/Anatomy/images/primaryxylem/images_c/Anat0144.jpg"
	tissslide2.Staining.Name = "Azul de metileno"

	tissrecord := tissuerecord.TissueRecord{
		Name:           "??? de un helecho",
		Notes:          "Corte 'y' de un '.' de un helecho, Pteridium sp. etc, etc, etc",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide1, tissslide2},
	}
	gorm_repository := repositories.NewGormTissueRecordRepository()
	tissrecord.ConfigureTissueRecord(gorm_repository)

	save_return := tissrecord.Save()

	if save_return == 0 {
		t.Errorf("got: %q, wanted %q", "zero", save_return)
	}
}

func TestTissueRecordRetrieve(t *testing.T) {
	// arrange
	tissslide1 := slide.Slide{Name: "img-10x"}
	tissslide2 := slide.Slide{Name: "img-200x"}
	tissrecord := tissuerecord.TissueRecord{
		Name:           "test retrive",
		Notes:          "'y' de un '.'",
		Taxonomicclass: "K:Any,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide1, tissslide2},
	}
	gorm_repository := repositories.NewGormTissueRecordRepository()
	tissrecord.ConfigureTissueRecord(gorm_repository)
	inserted_id := tissrecord.Save()

	// act
	_, status_code := tissrecord.GetById(inserted_id)

	// assert
	if status_code == 0 {
		t.Errorf("Not found record")
	}
}

func TestTissueRecordUpdate(t *testing.T) {
	// arrange
	tissslide1 := slide.Slide{Name: "img-10x"}
	tissslide2 := slide.Slide{Name: "img-200x"}
	tissrecord := tissuerecord.TissueRecord{
		Name:           "original name",
		Notes:          "'y' de un '.'",
		Taxonomicclass: "K:Any,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide1, tissslide2},
	}
	gorm_repository := repositories.NewGormTissueRecordRepository()
	tissrecord.ConfigureTissueRecord(gorm_repository)
	inserted_id := tissrecord.Save()

	tissslide3 := slide.Slide{Name: "img-10x"}
	tissslide4 := slide.Slide{Name: "img-200x"}
	tissrecord_to_update := tissuerecord.TissueRecord{
		Name:           "updated name",
		Notes:          "'y' de un '.'",
		Taxonomicclass: "K:Any,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide3, tissslide4},
	}

	// act
	tissrecord.Update(inserted_id, tissrecord_to_update)

	// assert
	retrieved, _ := tissrecord.GetById(inserted_id)
	if tissrecord.Name == retrieved.Name {
		t.Errorf("got: '%+v', wanted '%+v'", tissrecord.Name, tissrecord_to_update.Name)
	}
}

func TestTissueRecordDelete(t *testing.T) {
	// arrange
	tissslide1 := slide.Slide{Name: "img-10x"}
	tissslide2 := slide.Slide{Name: "img-200x"}
	tissrecord := tissuerecord.TissueRecord{
		Name:           "test retrive",
		Notes:          "'y' de un '.'",
		Taxonomicclass: "K:Any,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide1, tissslide2},
	}
	gorm_repository := repositories.NewGormTissueRecordRepository()
	tissrecord.ConfigureTissueRecord(gorm_repository)
	inserted_id := tissrecord.Save()

	// act
	tissrecord.Delete(inserted_id)

	// assert
	_, status_code := tissrecord.GetById(inserted_id)
	if status_code != 0 {
		t.Errorf("Not deleted")
	}
}
