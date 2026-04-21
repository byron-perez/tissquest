package collection

// RepositoryInterface defines persistence operations for collections.
type RepositoryInterface interface {
	Save(c *Collection) (uint, error)
	Retrieve(id uint) (*Collection, error)
	Update(id uint, c *Collection) error
	Delete(id uint) error
	List() ([]Collection, error)

	// Section operations
	CreateSection(s *Section) (uint, error)
	UpdateSection(id uint, s *Section) error
	DeleteSection(id uint) error
	ReorderSections(collectionID uint, positions map[uint]int) error

	// Assignment operations
	CreateAssignment(a *SectionAssignment) (uint, error)
	DeleteAssignment(id uint) error
	ReorderAssignments(sectionID uint, positions map[uint]int) error
}
