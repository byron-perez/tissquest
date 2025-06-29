package atlas

import (
	"errors"
	"time"
)

// Atlas represents an atlas entity in the domain
type Atlas struct {
	ID            uint
	Name          string
	Description   string
	Category      string
	TissueRecords []uint
	CreatedAt     time.Time
	UpdatedAt     time.Time
	repository    RepositoryInterface
}

// Error definitions
var (
	ErrEmptyName    = errors.New("atlas name cannot be empty")
	ErrNameTooLong  = errors.New("atlas name is too long (max 100 characters)")
	ErrNotFound     = errors.New("atlas not found")
	ErrInvalidInput = errors.New("invalid input data")
	ErrNoRepository = errors.New("no repository configured")
)

// ConfigureAtlas sets the repository for the atlas
func (a *Atlas) ConfigureAtlas(repository RepositoryInterface) {
	a.repository = repository
}

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

// Save persists the atlas to the repository
func (a *Atlas) Save() (uint, error) {
	if a.repository == nil {
		return 0, ErrNoRepository
	}

	if err := a.Validate(); err != nil {
		return 0, err
	}

	// Set timestamps
	now := time.Now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now

	return a.repository.Save(a), nil
}

// Load retrieves an atlas by ID from the repository
func (a *Atlas) Load(id uint) error {
	if a.repository == nil {
		return ErrNoRepository
	}

	atlas, err := a.repository.Retrieve(id)
	if err != nil {
		return err
	}

	// Copy data from retrieved atlas
	a.ID = atlas.ID
	a.Name = atlas.Name
	a.Description = atlas.Description
	a.Category = atlas.Category
	a.TissueRecords = atlas.TissueRecords
	a.CreatedAt = atlas.CreatedAt
	a.UpdatedAt = atlas.UpdatedAt

	return nil
}

// Update updates the atlas in the repository
func (a *Atlas) Update() error {
	if a.repository == nil {
		return ErrNoRepository
	}

	if err := a.Validate(); err != nil {
		return err
	}

	// Update timestamp
	a.UpdatedAt = time.Now()

	return a.repository.Update(a.ID, a)
}

// Delete removes the atlas from the repository
func (a *Atlas) Delete() error {
	if a.repository == nil {
		return ErrNoRepository
	}

	return a.repository.Delete(a.ID)
}

// AddTissueRecord adds a tissue record to the atlas
func (a *Atlas) AddTissueRecord(tissueRecordID uint) error {
	// Check if tissue record already exists
	for _, id := range a.TissueRecords {
		if id == tissueRecordID {
			return nil // Already exists
		}
	}

	// Add tissue record
	a.TissueRecords = append(a.TissueRecords, tissueRecordID)

	// Update the atlas if repository is configured
	if a.repository != nil {
		return a.Update()
	}

	return nil
}

// RemoveTissueRecord removes a tissue record from the atlas
func (a *Atlas) RemoveTissueRecord(tissueRecordID uint) error {
	for i, id := range a.TissueRecords {
		if id == tissueRecordID {
			// Remove the tissue record
			a.TissueRecords = append(a.TissueRecords[:i], a.TissueRecords[i+1:]...)

			// Update the atlas if repository is configured
			if a.repository != nil {
				return a.Update()
			}

			return nil
		}
	}

	return nil // Tissue record not found, no action needed
}

// GetTissueRecords returns the tissue records associated with this atlas
func (a *Atlas) GetTissueRecords() []uint {
	return a.TissueRecords
}

// NewAtlas creates a new Atlas instance with the given repository
func NewAtlas(repository RepositoryInterface) *Atlas {
	return &Atlas{
		repository: repository,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// FindAtlasByName searches for atlases by name using the repository
func FindAtlasByName(repository RepositoryInterface, name string) ([]Atlas, error) {
	if repository == nil {
		return nil, ErrNoRepository
	}

	return repository.FindByName(name)
}

// ListAllAtlases retrieves all atlases from the repository
func ListAllAtlases(repository RepositoryInterface) ([]Atlas, error) {
	if repository == nil {
		return nil, ErrNoRepository
	}

	return repository.List()
}

// ListAtlasesWithPagination retrieves atlases with pagination
func ListAtlasesWithPagination(repository RepositoryInterface, page, pageSize int) ([]Atlas, int64, error) {
	if repository == nil {
		return nil, 0, ErrNoRepository
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return repository.ListWithPagination(page, pageSize)
}
