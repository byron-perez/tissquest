package metacategory

import (
	"errors"
	"time"
)

// Metacategory represents a top-level grouping of categories.
// It allows arbitrary depth through ParentID self-reference.
// Example hierarchy:
//
//	método histológico (root)
//	  └─ tinción
//	      └─ tinción h&e
//	          └─ corte de hongo
type Metacategory struct {
	ID          uint
	Name        string
	Description string
	ParentID    *uint
	Parent      *Metacategory
	Children    []Metacategory
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrEmptyName      = errors.New("metacategory name cannot be empty")
	ErrCircularParent = errors.New("circular parent reference not allowed")
	ErrNotFound       = errors.New("metacategory not found")
)

func (m *Metacategory) Validate() error {
	if m.Name == "" {
		return ErrEmptyName
	}
	return nil
}

// IsRoot returns true if this is a top-level metacategory (no parent).
func (m *Metacategory) IsRoot() bool {
	return m.ParentID == nil
}

// HasChildren returns true if this metacategory has child metacategories.
func (m *Metacategory) HasChildren() bool {
	return len(m.Children) > 0
}
