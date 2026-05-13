package metacategory

type RepositoryInterface interface {
	// Save creates a new metacategory and returns its ID
	Save(m *Metacategory) (uint, error)

	// Retrieve gets a metacategory by ID
	Retrieve(id uint) (*Metacategory, error)

	// Update modifies an existing metacategory
	Update(id uint, m *Metacategory) error

	// Delete removes a metacategory
	Delete(id uint) error

	// List returns all root-level metacategories
	List() ([]Metacategory, error)

	// FindByParent returns all metacategories with the given parent ID
	FindByParent(parentID uint) ([]Metacategory, error)

	// FindRootMetacategories returns all top-level metacategories (ParentID is null)
	FindRootMetacategories() ([]Metacategory, error)

	// GetFullHierarchy retrieves a metacategory with all descendants loaded
	GetFullHierarchy(id uint) (*Metacategory, error)

	// ListWithChildren returns all root metacategories with their children preloaded
	ListWithChildren() ([]Metacategory, error)
}
