package tests

import (
	"os"
	"strings"
	"testing"
	"unicode"

	"mcba/tissquest/internal/core/collection"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/migration"
	"mcba/tissquest/internal/persistence/repositories"

	"pgregory.net/rapid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Test DB helpers ---

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	f, err := os.CreateTemp("", "tissquest-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	db, err := gorm.Open(sqlite.Open(f.Name()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&migration.CollectionModel{},
		&migration.CollectionSectionModel{},
		&migration.CollectionSectionAssignmentModel{},
		&migration.TissueRecordModel{},
		&migration.TaxonModel{},
		&migration.PreparationModel{},
		&migration.SlideModel{},
		&migration.SlideImageVariantModel{},
		&migration.CategoryModel{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func newTestRepo(t *testing.T) (*repositories.GormCollectionRepository, *gorm.DB) {
	t.Helper()
	db := newTestDB(t)
	return repositories.NewGormCollectionRepositoryWithDB(db), db
}

// --- Generators ---

func genWhitespaceString(t *rapid.T) string {
	n := rapid.IntRange(0, 20).Draw(t, "len")
	ws := []rune{' ', '\t', '\n', '\r'}
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteRune(ws[rapid.IntRange(0, len(ws)-1).Draw(t, "ws")])
	}
	return b.String()
}

func genValidName(t *rapid.T) string {
	s := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9 ]{0,50}`).Draw(t, "name")
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

func genLongName(t *rapid.T) string {
	n := rapid.IntRange(201, 300).Draw(t, "len")
	return strings.Repeat("a", n)
}

func genInvalidType(t *rapid.T) string {
	valid := map[string]bool{"atlas": true, "database": true, "reference": true, "other": true}
	s := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "type")
	if valid[s] {
		return s + "x"
	}
	return s
}

func isWhitespaceOnly(s string) bool {
	return strings.TrimFunc(s, unicode.IsSpace) == ""
}

func findSectionByID(col *collection.Collection, id uint) *collection.Section {
	for i := range col.Sections {
		if col.Sections[i].ID == id {
			return &col.Sections[i]
		}
		for j := range col.Sections[i].Subsections {
			if col.Sections[i].Subsections[j].ID == id {
				return &col.Sections[i].Subsections[j]
			}
		}
	}
	return nil
}

// --- Property 1: Whitespace and empty names are rejected ---
// Feature: collection-builder, Property 1: Whitespace and empty names are rejected

func TestProperty1_WhitespaceCollectionNameRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		name := genWhitespaceString(rt)
		if !isWhitespaceOnly(name) {
			return
		}
		c := &collection.Collection{Name: name, Type: collection.CollectionTypeAtlas}
		if err := c.Validate(); err != collection.ErrEmptyName {
			rt.Fatalf("expected ErrEmptyName for name %q, got %v", name, err)
		}
	})
}

func TestProperty1_WhitespaceSectionNameRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		name := genWhitespaceString(rt)
		if !isWhitespaceOnly(name) {
			return
		}
		s := &collection.Section{Name: name}
		if err := s.Validate(); err != collection.ErrEmptyName {
			rt.Fatalf("expected ErrEmptyName for section name %q, got %v", name, err)
		}
	})
}

// --- Property 2: Name length boundary enforcement ---
// Feature: collection-builder, Property 2: Name length boundary enforcement

func TestProperty2_NameTooLong(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		name := genLongName(rt)
		c := &collection.Collection{Name: name, Type: collection.CollectionTypeAtlas}
		if err := c.Validate(); err != collection.ErrNameTooLong {
			rt.Fatalf("expected ErrNameTooLong for name of length %d, got %v", len(name), err)
		}
	})
}

// --- Property 3: Collection type enum enforcement ---
// Feature: collection-builder, Property 3: Collection type enum enforcement

func TestProperty3_InvalidTypeRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		typ := genInvalidType(rt)
		c := &collection.Collection{Name: "Valid Name", Type: collection.CollectionType(typ)}
		if err := c.Validate(); err != collection.ErrInvalidType {
			rt.Fatalf("expected ErrInvalidType for type %q, got %v", typ, err)
		}
	})
}

// --- Property 4: Collection metadata round-trip ---
// Feature: collection-builder, Property 4: Collection metadata round-trip

func TestProperty4_MetadataRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, _ := newTestRepo(t)
		name := genValidName(rt)
		desc := rapid.String().Draw(rt, "desc")
		goals := rapid.String().Draw(rt, "goals")
		authors := rapid.String().Draw(rt, "authors")

		c := &collection.Collection{
			Name:        name,
			Description: desc,
			Goals:       goals,
			Type:        collection.CollectionTypeAtlas,
			Authors:     authors,
		}
		id, err := repo.Save(c)
		if err != nil {
			rt.Fatalf("Save failed: %v", err)
		}
		got, err := repo.Retrieve(id)
		if err != nil {
			rt.Fatalf("Retrieve failed: %v", err)
		}
		if got.Name != name {
			rt.Fatalf("Name mismatch: got %q, want %q", got.Name, name)
		}
		if got.Description != desc {
			rt.Fatalf("Description mismatch: got %q, want %q", got.Description, desc)
		}
		if got.Goals != goals {
			rt.Fatalf("Goals mismatch: got %q, want %q", got.Goals, goals)
		}
		if got.Authors != authors {
			rt.Fatalf("Authors mismatch: got %q, want %q", got.Authors, authors)
		}
	})
}

// --- Property 5: Section creation assigns next position ---
// Feature: collection-builder, Property 5: Section creation assigns next position

func TestProperty5_SectionCreationPosition(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, _ := newTestRepo(t)
		colID, _ := repo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})

		n := rapid.IntRange(0, 10).Draw(rt, "existing")
		for i := 0; i < n; i++ {
			repo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})
		}

		secID, err := repo.CreateSection(&collection.Section{CollectionID: colID, Name: "New"})
		if err != nil {
			rt.Fatalf("CreateSection failed: %v", err)
		}

		got, _ := repo.Retrieve(colID)
		found := findSectionByID(got, secID)
		if found == nil {
			rt.Fatalf("new section not found")
		}
		if found.Position != n+1 {
			rt.Fatalf("expected position %d, got %d", n+1, found.Position)
		}
	})
}

// --- Property 6: Reorder persists new positions ---
// Feature: collection-builder, Property 6: Reorder persists new positions

func TestProperty6_ReorderSectionsPersists(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, _ := newTestRepo(t)
		colID, _ := repo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})

		n := rapid.IntRange(2, 6).Draw(rt, "n")
		ids := make([]uint, n)
		for i := 0; i < n; i++ {
			id, _ := repo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})
			ids[i] = id
		}

		// Reverse the positions
		positions := make(map[uint]int)
		for i, id := range ids {
			positions[id] = n - i
		}
		if err := repo.ReorderSections(colID, positions); err != nil {
			rt.Fatalf("ReorderSections failed: %v", err)
		}

		got, _ := repo.Retrieve(colID)
		for _, sec := range got.Sections {
			if want, ok := positions[sec.ID]; ok {
				if sec.Position != want {
					rt.Fatalf("section %d: expected position %d, got %d", sec.ID, want, sec.Position)
				}
			}
		}
	})
}

// --- Property 7: Section deletion removes all assignments ---
// Feature: collection-builder, Property 7: Section deletion removes all assignments

func TestProperty7_SectionDeletionClearsAssignments(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, db := newTestRepo(t)
		colID, _ := repo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})
		secID, _ := repo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})

		n := rapid.IntRange(0, 20).Draw(rt, "assignments")
		for i := 0; i < n; i++ {
			repo.CreateAssignment(&collection.SectionAssignment{
				SectionID:      secID,
				TissueRecordID: uint(i + 1),
			})
		}

		if err := repo.DeleteSection(secID); err != nil {
			rt.Fatalf("DeleteSection failed: %v", err)
		}

		var count int64
		db.Model(&migration.CollectionSectionAssignmentModel{}).
			Where("section_id = ?", secID).Count(&count)
		if count != 0 {
			rt.Fatalf("expected 0 assignments after section delete, got %d", count)
		}
	})
}

// --- Property 8: Assignment creation appends at end ---
// Feature: collection-builder, Property 8: Assignment creation appends at end

func TestProperty8_AssignmentAppendsAtEnd(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, _ := newTestRepo(t)
		colID, _ := repo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})
		secID, _ := repo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})

		n := rapid.IntRange(0, 10).Draw(rt, "existing")
		for i := 0; i < n; i++ {
			repo.CreateAssignment(&collection.SectionAssignment{
				SectionID:      secID,
				TissueRecordID: uint(i + 1),
			})
		}

		newID, err := repo.CreateAssignment(&collection.SectionAssignment{
			SectionID:      secID,
			TissueRecordID: uint(n + 100),
		})
		if err != nil {
			rt.Fatalf("CreateAssignment failed: %v", err)
		}

		got, _ := repo.Retrieve(colID)
		sec := findSectionByID(got, secID)
		if sec == nil {
			rt.Fatalf("section not found")
		}
		var found *collection.SectionAssignment
		for i := range sec.Assignments {
			if sec.Assignments[i].ID == newID {
				found = &sec.Assignments[i]
				break
			}
		}
		if found == nil {
			rt.Fatalf("new assignment not found")
		}
		if found.Position != n+1 {
			rt.Fatalf("expected position %d, got %d", n+1, found.Position)
		}
	})
}

// --- Property 9: Duplicate assignment rejection ---
// Feature: collection-builder, Property 9: Duplicate assignment rejection

func TestProperty9_DuplicateAssignmentRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, _ := newTestRepo(t)
		colID, _ := repo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})
		secID, _ := repo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})

		trID := uint(42)
		if _, err := repo.CreateAssignment(&collection.SectionAssignment{SectionID: secID, TissueRecordID: trID}); err != nil {
			rt.Fatalf("first assignment failed: %v", err)
		}

		if _, err := repo.CreateAssignment(&collection.SectionAssignment{SectionID: secID, TissueRecordID: trID}); err != collection.ErrDuplicateAssignment {
			rt.Fatalf("expected ErrDuplicateAssignment, got %v", err)
		}

		got, _ := repo.Retrieve(colID)
		sec := findSectionByID(got, secID)
		if sec == nil {
			rt.Fatalf("section not found")
		}
		if len(sec.Assignments) != 1 {
			rt.Fatalf("expected 1 assignment, got %d", len(sec.Assignments))
		}
	})
}

// --- Property 10: Assignment removal resequences positions ---
// Feature: collection-builder, Property 10: Assignment removal resequences positions

func TestProperty10_AssignmentRemovalResequences(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, _ := newTestRepo(t)
		colID, _ := repo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})
		secID, _ := repo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})

		n := rapid.IntRange(2, 10).Draw(rt, "n")
		ids := make([]uint, n)
		for i := 0; i < n; i++ {
			id, _ := repo.CreateAssignment(&collection.SectionAssignment{
				SectionID:      secID,
				TissueRecordID: uint(i + 1),
			})
			ids[i] = id
		}

		removeIdx := rapid.IntRange(0, n-1).Draw(rt, "removeIdx")
		if err := repo.DeleteAssignment(ids[removeIdx]); err != nil {
			rt.Fatalf("DeleteAssignment failed: %v", err)
		}

		got, _ := repo.Retrieve(colID)
		sec := findSectionByID(got, secID)
		if sec == nil {
			rt.Fatalf("section not found")
		}
		if len(sec.Assignments) != n-1 {
			rt.Fatalf("expected %d assignments, got %d", n-1, len(sec.Assignments))
		}
		for i, a := range sec.Assignments {
			if a.Position != i+1 {
				rt.Fatalf("position gap at index %d: got %d, want %d", i, a.Position, i+1)
			}
		}
	})
}

// --- Property 11: Search returns only matching records ---
// Feature: collection-builder, Property 11: Search returns only matching records

func TestProperty11_SearchReturnsOnlyMatching(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Use an empty TR repo — search on empty pool should always return empty
		// and never return non-matching results
		_, db := newTestRepo(t)
		trRepo := repositories.NewGormTissueRecordRepositoryWithDB(db)
		colRepo := repositories.NewGormCollectionRepositoryWithDB(db)
		svc := newCollectionSvc(colRepo, trRepo)

		query := rapid.StringMatching(`[a-z]{2,5}`).Draw(rt, "query")
		results, err := svc.SearchTissueRecords(query)
		if err != nil {
			rt.Fatalf("SearchTissueRecords failed: %v", err)
		}
		q := strings.ToLower(query)
		for _, r := range results {
			nameMatch := strings.Contains(strings.ToLower(r.Name), q)
			taxonMatch := r.Taxon != nil && strings.Contains(strings.ToLower(r.Taxon.Name), q)
			if !nameMatch && !taxonMatch {
				rt.Fatalf("result %q does not match query %q", r.Name, query)
			}
		}
	})
}

// --- Property 12: Inline creation persists record and creates assignment ---
// Feature: collection-builder, Property 12: Inline creation persists record and creates assignment

func TestProperty12_InlineCreationPersistsAndAssigns(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		colRepo, db := newTestRepo(t)
		trRepo := repositories.NewGormTissueRecordRepositoryWithDB(db)
		svc := newCollectionSvc(colRepo, trRepo)

		colID, _ := colRepo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})
		secID, _ := colRepo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})

		name := genValidName(rt)
		tr := &tissuerecord.TissueRecord{Name: name}
		if err := svc.CreateTissueRecordAndAssign(tr, secID); err != nil {
			rt.Fatalf("CreateTissueRecordAndAssign failed: %v", err)
		}
		if tr.ID == 0 {
			rt.Fatalf("tissue record ID not set after creation")
		}

		got, _ := colRepo.Retrieve(colID)
		sec := findSectionByID(got, secID)
		if sec == nil {
			rt.Fatalf("section not found")
		}
		found := false
		for _, a := range sec.Assignments {
			if a.TissueRecordID == tr.ID {
				found = true
				break
			}
		}
		if !found {
			rt.Fatalf("assignment for TR %d not found in section", tr.ID)
		}
	})
}

// --- Property 13: Collection deletion cascades ---
// Feature: collection-builder, Property 13: Collection deletion cascades

func TestProperty13_CollectionDeletionCascades(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, db := newTestRepo(t)
		colID, _ := repo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})

		nSections := rapid.IntRange(1, 5).Draw(rt, "sections")
		for i := 0; i < nSections; i++ {
			secID, _ := repo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})
			nAssign := rapid.IntRange(0, 5).Draw(rt, "assigns")
			for j := 0; j < nAssign; j++ {
				repo.CreateAssignment(&collection.SectionAssignment{
					SectionID:      secID,
					TissueRecordID: uint(j + 1),
				})
			}
		}

		if err := repo.Delete(colID); err != nil {
			rt.Fatalf("Delete failed: %v", err)
		}

		var secCount int64
		db.Model(&migration.CollectionSectionModel{}).Where("collection_id = ?", colID).Count(&secCount)
		if secCount != 0 {
			rt.Fatalf("expected 0 sections after collection delete, got %d", secCount)
		}
	})
}

// --- Property 14: Tissue record deletion removes all assignments ---
// Feature: collection-builder, Property 14: Tissue record deletion removes all assignments

func TestProperty14_TissueRecordDeletionRemovesAssignments(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, db := newTestRepo(t)
		colID, _ := repo.Save(&collection.Collection{Name: "C", Type: collection.CollectionTypeAtlas})

		sharedTRID := uint(999)
		nSections := rapid.IntRange(1, 4).Draw(rt, "sections")
		for i := 0; i < nSections; i++ {
			secID, _ := repo.CreateSection(&collection.Section{CollectionID: colID, Name: "S"})
			repo.CreateAssignment(&collection.SectionAssignment{
				SectionID:      secID,
				TissueRecordID: sharedTRID,
			})
		}

		// Simulate TR deletion cascading to assignments
		db.Where("tissue_record_id = ?", sharedTRID).Delete(&migration.CollectionSectionAssignmentModel{})

		var count int64
		db.Model(&migration.CollectionSectionAssignmentModel{}).
			Where("tissue_record_id = ?", sharedTRID).Count(&count)
		if count != 0 {
			rt.Fatalf("expected 0 assignments after TR delete, got %d", count)
		}
	})
}

// --- Property 15: Collection list rendering includes required fields ---
// Feature: collection-builder, Property 15: Collection list rendering includes required fields

func TestProperty15_CollectionListHasRequiredFields(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo, _ := newTestRepo(t)
		n := rapid.IntRange(1, 5).Draw(rt, "n")
		names := make([]string, n)
		for i := 0; i < n; i++ {
			name := genValidName(rt)
			names[i] = name
			repo.Save(&collection.Collection{Name: name, Type: collection.CollectionTypeAtlas})
		}

		cols, err := repo.List()
		if err != nil {
			rt.Fatalf("List failed: %v", err)
		}
		if len(cols) < n {
			rt.Fatalf("expected at least %d collections, got %d", n, len(cols))
		}
		for _, name := range names {
			found := false
			for _, col := range cols {
				if col.Name == name {
					if col.Type == "" {
						rt.Fatalf("collection %q has empty type", name)
					}
					if col.CreatedAt.IsZero() {
						rt.Fatalf("collection %q has zero CreatedAt", name)
					}
					found = true
					break
				}
			}
			if !found {
				rt.Fatalf("collection %q not found in list", name)
			}
		}
	})
}
