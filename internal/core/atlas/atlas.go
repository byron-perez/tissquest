package atlas

import (
	"errors"
	"time"
)

// Atlas represents an atlas entity in the domain
type Atlas struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	TissueRecords []uint    `json:"tissue_records"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Error definitions
var (
	ErrEmptyName    = errors.New("atlas name cannot be empty")
	ErrNameTooLong  = errors.New("atlas name is too long (max 100 characters)")
	ErrNotFound     = errors.New("atlas not found")
	ErrInvalidInput = errors.New("invalid input data")
)

// Validate validates the atlas data
func (a *Atlas) Validate() error {
	if a.Name == "" {
		return ErrEmptyName
	}

	if len(a.Name) > 100 {
		return ErrNameTooLong
	}

	return nil
}

// AddTissueRecord adds a tissue record to the atlas
func (a *Atlas) AddTissueRecord(tissueRecordID uint) {
	// Check if tissue record already exists
	for _, id := range a.TissueRecords {
		if id == tissueRecordID {
			return // Already exists
		}
	}

	// Add tissue record
	a.TissueRecords = append(a.TissueRecords, tissueRecordID)
}

// RemoveTissueRecord removes a tissue record from the atlas
func (a *Atlas) RemoveTissueRecord(tissueRecordID uint) {
	for i, id := range a.TissueRecords {
		if id == tissueRecordID {
			// Remove the tissue record
			a.TissueRecords = append(a.TissueRecords[:i], a.TissueRecords[i+1:]...)
			return
		}
	}
}

// GetTissueRecords returns the tissue records associated with this atlas
func (a *Atlas) GetTissueRecords() []uint {
	return a.TissueRecords
}
