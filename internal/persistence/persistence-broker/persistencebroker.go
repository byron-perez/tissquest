package persistencebroker

import (
	"mcba/tissquest/internal/core/slide"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PersistenceBroker struct {
	singleInstance struct{}
	connections    []struct{}
}

func (pb *PersistenceBroker) NewPersistenceBroker(objectinstance struct{}) {
	pb.singleInstance = objectinstance
}

type TissueRecordModel struct {
	gorm.Model
	Name           string
	Notes          string
	Taxonomicclass string
	Slides         []slide.Slide
}

func (pb *PersistenceBroker) SaveObject() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&TissueRecordModel{})

	// db.Create(&TissueRecordModel{
	// 	Name:           pb.singleInstance.Name,
	// 	Notes:          pb.singleInstance.Notes,
	// 	Taxonomicclass: pb.singleInstance.Taxonomicclass,
	// 	Slides:         pb.singleInstance.Slides,
	// })

}
