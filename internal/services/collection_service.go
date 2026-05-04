package services

import (
	"strings"

	"mcba/tissquest/internal/core/collection"
	"mcba/tissquest/internal/core/tissuerecord"
)

// CollectionService handles business logic for collections.
type CollectionService struct {
	repo   collection.RepositoryInterface
	trRepo tissuerecord.RepositoryInterface
}

// NewCollectionService creates a new CollectionService.
func NewCollectionService(repo collection.RepositoryInterface, trRepo tissuerecord.RepositoryInterface) *CollectionService {
	return &CollectionService{repo: repo, trRepo: trRepo}
}

// CreateCollection validates and persists a new collection.
func (s *CollectionService) CreateCollection(c *collection.Collection) (uint, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}
	return s.repo.Save(c)
}

// GetCollection retrieves a collection by ID (with sections/assignments).
func (s *CollectionService) GetCollection(id uint) (*collection.Collection, error) {
	return s.repo.Retrieve(id)
}

// UpdateCollection validates and updates a collection's metadata.
func (s *CollectionService) UpdateCollection(id uint, c *collection.Collection) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return s.repo.Update(id, c)
}

// DeleteCollection deletes a collection and cascades to sections/assignments.
func (s *CollectionService) DeleteCollection(id uint) error {
	return s.repo.Delete(id)
}

// ListCollections returns all collections.
func (s *CollectionService) ListCollections() ([]collection.Collection, error) {
	return s.repo.List()
}

// CreateSection creates a new section in a collection.
func (s *CollectionService) CreateSection(collectionID uint, name string, parentID *uint) (uint, error) {
	sec := &collection.Section{
		CollectionID: collectionID,
		ParentID:     parentID,
		Name:         name,
	}
	if err := sec.Validate(); err != nil {
		return 0, err
	}
	return s.repo.CreateSection(sec)
}

// RenameSection renames an existing section.
func (s *CollectionService) RenameSection(sectionID uint, name string) error {
	sec := &collection.Section{Name: name}
	if err := sec.Validate(); err != nil {
		return err
	}
	return s.repo.UpdateSection(sectionID, sec)
}

// DeleteSection deletes a section and all its assignments.
func (s *CollectionService) DeleteSection(sectionID uint) error {
	return s.repo.DeleteSection(sectionID)
}

// ReorderSections persists new positions for sections within a collection.
func (s *CollectionService) ReorderSections(collectionID uint, positions map[uint]int) error {
	return s.repo.ReorderSections(collectionID, positions)
}

// AssignTissueRecord assigns a tissue record to a section.
// Returns ErrDuplicateAssignment if already assigned.
func (s *CollectionService) AssignTissueRecord(sectionID, tissueRecordID uint) (uint, error) {
	a := &collection.SectionAssignment{
		SectionID:      sectionID,
		TissueRecordID: tissueRecordID,
	}
	return s.repo.CreateAssignment(a)
}

// RemoveAssignment removes a section assignment and resequences positions.
func (s *CollectionService) RemoveAssignment(assignmentID uint) error {
	return s.repo.DeleteAssignment(assignmentID)
}

// ReorderAssignments persists new positions for assignments within a section.
func (s *CollectionService) ReorderAssignments(sectionID uint, positions map[uint]int) error {
	return s.repo.ReorderAssignments(sectionID, positions)
}

// SearchTissueRecords performs a case-insensitive substring search on tissue record name and taxon name.
func (s *CollectionService) SearchTissueRecords(query string) ([]tissuerecord.TissueRecord, error) {
	if s.trRepo == nil {
		return nil, nil
	}
	// Fetch all records (paginated with a large limit for search)
	records, _, err := s.trRepo.List(1, 1000)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var results []tissuerecord.TissueRecord
	for _, r := range records {
		if strings.Contains(strings.ToLower(r.Name), q) {
			results = append(results, r)
			continue
		}
		if r.Taxon != nil && strings.Contains(strings.ToLower(r.Taxon.Name), q) {
			results = append(results, r)
		}
	}
	return results, nil
}

// CreateTissueRecordAndAssign persists a new tissue record and creates a section assignment.
func (s *CollectionService) CreateTissueRecordAndAssign(tr *tissuerecord.TissueRecord, sectionID uint) error {
	if s.trRepo == nil {
		return nil
	}
	id := s.trRepo.Save(tr)
	tr.ID = id
	_, err := s.AssignTissueRecord(sectionID, id)
	return err
}
