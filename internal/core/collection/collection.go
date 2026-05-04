package collection

import (
	"errors"
	"strings"
	"time"
)

// CollectionType classifies a collection's intended purpose.
type CollectionType string

const (
	CollectionTypeAtlas     CollectionType = "atlas"
	CollectionTypeDatabase  CollectionType = "database"
	CollectionTypeReference CollectionType = "reference"
	CollectionTypeOther     CollectionType = "other"
)

var validTypes = map[CollectionType]struct{}{
	CollectionTypeAtlas:     {},
	CollectionTypeDatabase:  {},
	CollectionTypeReference: {},
	CollectionTypeOther:     {},
}

// Domain errors
var (
	ErrEmptyName           = errors.New("name cannot be empty or whitespace-only")
	ErrNameTooLong         = errors.New("name exceeds maximum length of 200 characters")
	ErrInvalidType         = errors.New("invalid collection type; must be atlas, database, reference, or other")
	ErrNotFound            = errors.New("collection not found")
	ErrDuplicateAssignment = errors.New("tissue record is already assigned to this section")
	ErrMaxDepthExceeded    = errors.New("maximum nesting depth of 2 levels exceeded")
)

// Collection is a named, curated grouping of tissue records.
type Collection struct {
	ID          uint
	Name        string
	Description string
	Goals       string
	Type        CollectionType
	Authors     string
	Sections    []Section
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate validates the collection's fields.
func (c *Collection) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrEmptyName
	}
	if len(c.Name) > 200 {
		return ErrNameTooLong
	}
	if c.Type != "" {
		if _, ok := validTypes[c.Type]; !ok {
			return ErrInvalidType
		}
	}
	return nil
}

// Section is a named, ordered subdivision of a Collection.
type Section struct {
	ID           uint
	CollectionID uint
	ParentID     *uint
	Name         string
	Position     int
	Assignments  []SectionAssignment
	Subsections  []Section
}

// Validate validates the section's fields.
func (s *Section) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return ErrEmptyName
	}
	return nil
}

// SectionAssignment links a TissueRecord to a Section with an explicit position.
type SectionAssignment struct {
	ID               uint
	SectionID        uint
	TissueRecordID   uint
	TissueRecordName string // denormalized for display — populated by the repository
	Position         int
}
