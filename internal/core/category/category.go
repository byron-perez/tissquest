package category

import (
	"errors"
	"time"
)

type CategoryType string

const (
	CategoryOrgan   CategoryType = "organ"
	CategorySpecies CategoryType = "species"
	CategoryTissue  CategoryType = "tissue"
	CategoryStain   CategoryType = "stain"
	CategoryCustom  CategoryType = "custom"
)

type Category struct {
	ID              uint         `json:"id"`
	Name            string       `json:"name"`
	Type            CategoryType `json:"type"`
	Description     string       `json:"description"`
	ParentID        *uint        `json:"parent_id,omitempty"`
	TissueRecordIDs []uint       `json:"tissue_record_ids"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

var (
	ErrEmptyName        = errors.New("category name cannot be empty")
	ErrInvalidType      = errors.New("invalid category type")
	ErrCircularParent   = errors.New("circular parent reference not allowed")
)

func (c *Category) Validate() error {
	if c.Name == "" {
		return ErrEmptyName
	}
	
	validTypes := []CategoryType{CategoryOrgan, CategorySpecies, CategoryTissue, CategoryStain, CategoryCustom}
	valid := false
	for _, t := range validTypes {
		if c.Type == t {
			valid = true
			break
		}
	}
	if !valid {
		return ErrInvalidType
	}
	
	if c.ParentID != nil && *c.ParentID == c.ID {
		return ErrCircularParent
	}
	
	return nil
}

func (c *Category) AddTissueRecord(tissueRecordID uint) {
	for _, id := range c.TissueRecordIDs {
		if id == tissueRecordID {
			return
		}
	}
	c.TissueRecordIDs = append(c.TissueRecordIDs, tissueRecordID)
}

func (c *Category) RemoveTissueRecord(tissueRecordID uint) {
	for i, id := range c.TissueRecordIDs {
		if id == tissueRecordID {
			c.TissueRecordIDs = append(c.TissueRecordIDs[:i], c.TissueRecordIDs[i+1:]...)
			return
		}
	}
}

func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}