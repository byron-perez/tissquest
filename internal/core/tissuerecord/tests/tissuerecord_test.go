package tests

import (
	"log"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/migration"
	"mcba/tissquest/internal/persistence/repositories"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../../../../.env")
	if err != nil {
		log.Fatal(err.Error())
	}
	migration.RunMigration()
	m.Run()
}

func TestTissueRecordSave(t *testing.T) {
	repo := repositories.NewGormTissueRecordRepository()

	tr := &tissuerecord.TissueRecord{
		Name:  "Helecho sp.",
		Notes: "Corte transversal de un helecho",
	}

	id := repo.Save(tr)
	if id == 0 {
		t.Errorf("expected non-zero ID after save, got 0")
	}
}

func TestTissueRecordRetrieve(t *testing.T) {
	repo := repositories.NewGormTissueRecordRepository()

	tr := &tissuerecord.TissueRecord{
		Name:  "test retrieve",
		Notes: "notes for retrieve test",
	}
	id := repo.Save(tr)

	retrieved, statusCode := repo.Retrieve(id)
	if statusCode == 0 {
		t.Errorf("expected non-zero status code, record not found")
	}
	if retrieved.Name != tr.Name {
		t.Errorf("got name %q, want %q", retrieved.Name, tr.Name)
	}
}

func TestTissueRecordUpdate(t *testing.T) {
	repo := repositories.NewGormTissueRecordRepository()

	original := &tissuerecord.TissueRecord{
		Name:  "original name",
		Notes: "original notes",
	}
	id := repo.Save(original)

	updated := &tissuerecord.TissueRecord{
		Name:  "updated name",
		Notes: "updated notes",
	}
	repo.Update(id, updated)

	retrieved, _ := repo.Retrieve(id)
	if retrieved.Name != updated.Name {
		t.Errorf("got name %q, want %q", retrieved.Name, updated.Name)
	}
}

func TestTissueRecordDelete(t *testing.T) {
	repo := repositories.NewGormTissueRecordRepository()

	tr := &tissuerecord.TissueRecord{
		Name:  "to be deleted",
		Notes: "delete test",
	}
	id := repo.Save(tr)

	repo.Delete(id)

	_, statusCode := repo.Retrieve(id)
	if statusCode != 0 {
		t.Errorf("expected record to be deleted, but it was found (status %d)", statusCode)
	}
}
