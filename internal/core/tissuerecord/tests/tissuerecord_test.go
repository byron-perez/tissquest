package tests

import (
	"fmt"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
	"reflect"
	"testing"
)

func TestTissueRecord(t *testing.T) {
	tissslide1 := slide.Slide{Name: "img-40x"}
	tissslide2 := slide.Slide{Name: "img-100x"}
	tissfile1 := tissuerecord.TissueRecord{
		Name:           "Esporofito del helecho",
		Notes:          "Placa de un corte transversal del esporfito de un helecho, Pteridium sp.",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide1, tissslide2},
	}

	fmt.Println("#1")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile1.Name)
	fmt.Println("Explicación: " + tissfile1.Notes)
	fmt.Println("Taxonomía: " + tissfile1.Taxonomicclass)
	for _, slide := range tissfile1.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
	fmt.Println("------------------------------")
	tissslide3 := slide.Slide{Name: "img-10x"}
	tissslide4 := slide.Slide{Name: "img-40"}
	tissfile2 := tissuerecord.TissueRecord{
		Name:           "Xilema del helecho",
		Notes:          "Placa de un corte transversal del xilema de un helecho, Pteridium sp.",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide3, tissslide4},
	}

	fmt.Println("#2")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile2.Name)
	fmt.Println("Explicación: " + tissfile2.Notes)
	fmt.Println("Taxonomía: " + tissfile2.Taxonomicclass)
	for _, slide := range tissfile2.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
	fmt.Println("------------------------------")
	tissslide21 := slide.Slide{Name: "img-10x"}
	tissslide22 := slide.Slide{Name: "img-40"}
	tissfile3 := tissuerecord.TissueRecord{
		Name:           "Elementos cribosos de un helecho",
		Notes:          "Placa de un corte transversal del floema de un helecho, Pteridium sp.",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide21, tissslide22},
	}

	fmt.Println("#3")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile3.Name)
	fmt.Println("Explicación: " + tissfile3.Notes)
	fmt.Println("Taxonomía: " + tissfile3.Taxonomicclass)
	for _, slide := range tissfile3.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
	fmt.Println("------------------------------")
	tissslide41 := slide.Slide{Name: "img-10x"}
	tissslide42 := slide.Slide{Name: "img-40"}
	tissfile4 := tissuerecord.TissueRecord{
		Name:           "Fronda ramificada de un helecho",
		Notes:          "Placa de un corte longitudinal del esporofito de un helecho, Pteridium sp.",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide41, tissslide42},
	}

	fmt.Println("#4")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile4.Name)
	fmt.Println("Explicación: " + tissfile4.Notes)
	fmt.Println("Taxonomía: " + tissfile4.Taxonomicclass)
	for _, slide := range tissfile4.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
	fmt.Println("------------------------------")
	tissslide51 := slide.Slide{Name: "img-10x"}
	tissslide52 := slide.Slide{Name: "img-200x"}
	tissfile5 := tissuerecord.TissueRecord{
		Name:           "Endodermis de la raíz de un helecho",
		Notes:          "Placa de un corte transversal de la raíz de un helecho, Pteridium sp.",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide51, tissslide52},
	}

	fmt.Println("#5")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile5.Name)
	fmt.Println("Explicación: " + tissfile5.Notes)
	fmt.Println("Taxonomía: " + tissfile5.Taxonomicclass)
	for _, slide := range tissfile5.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
	fmt.Println("------------------------------")
	tissslide61 := slide.Slide{Name: "img-10x"}
	tissslide62 := slide.Slide{Name: "img-200x"}
	tissfile6 := tissuerecord.TissueRecord{
		Name:           "Rizoma de un helecho",
		Notes:          "Corte transversal del rizoma de un helecho, Pteridium sp. se puede ver la endodermis en color azul",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide61, tissslide62},
	}

	fmt.Println("#6")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile6.Name)
	fmt.Println("Explicación: " + tissfile6.Notes)
	fmt.Println("Taxonomía: " + tissfile6.Taxonomicclass)
	for _, slide := range tissfile6.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
	fmt.Println("------------------------------")
	tissslide71 := slide.Slide{Name: "img-10x"}
	tissslide72 := slide.Slide{Name: "img-200x"}
	tissfile7 := tissuerecord.TissueRecord{
		Name:           "Tejido lignificado de un helecho",
		Notes:          "Corte transversal del brote de un helecho, Pteridium sp. se puede ver las células de exterior del xilema de un diámetro de 10 nanómetros.",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide71, tissslide72},
	}

	fmt.Println("#7")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile7.Name)
	fmt.Println("Explicación: " + tissfile7.Notes)
	fmt.Println("Taxonomía: " + tissfile7.Taxonomicclass)
	for _, slide := range tissfile7.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
	fmt.Println("------------------------------")
	tissslide81 := slide.Slide{Name: "img-10x"}
	tissslide82 := slide.Slide{Name: "img-200x"}
	tissfile8 := tissuerecord.TissueRecord{
		Name:           "Esclerénquima de un helecho",
		Notes:          "Corte longitudinal de una fronda de un helecho, Pteridium sp. se puede ver las células del colénquima que tienen un diámetro de 10 amstrongs.",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide81, tissslide82},
	}

	fmt.Println("#8")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile8.Name)
	fmt.Println("Explicación: " + tissfile8.Notes)
	fmt.Println("Taxonomía: " + tissfile8.Taxonomicclass)
	for _, slide := range tissfile8.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
	fmt.Println("------------------------------")
	tissslide91 := slide.Slide{Name: "img-10x"}
	tissslide92 := slide.Slide{Name: "img-200x"}
	tissfile9 := tissuerecord.TissueRecord{
		Name:           "??? de un helecho",
		Notes:          "Corte 'y' de un '.' de un helecho, Pteridium sp. etc, etc, etc",
		Taxonomicclass: "K:Plantae,Cld:Tracheophytes,D:Polypodiophyta,Cls:Polypodiopsida",
		Slides:         []slide.Slide{tissslide91, tissslide92},
	}

	fmt.Println("#9")
	fmt.Println("Aquí su tejido:")
	fmt.Println("Nombre: " + tissfile9.Name)
	fmt.Println("Explicación: " + tissfile9.Notes)
	fmt.Println("Taxonomía: " + tissfile9.Taxonomicclass)
	for _, slide := range tissfile9.Slides {
		fmt.Println("Imágenes: " + slide.Name)
	}
}

func TestTissueRecordSave(t *testing.T) {
	tissslide1 := slide.Slide{Name: "img-10x"}
	tissslide2 := slide.Slide{Name: "img-200x"}
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

	retrieved := tissrecord.GetById(inserted_id)
	if reflect.DeepEqual(tissrecord, retrieved) {
		t.Errorf("got: %q, wanted %q", retrieved, tissrecord)
	}
}
