package atlas

type RepositoryInterface interface {
	// Basic CRUD operations
	Save(a *Atlas) uint
	Retrieve(id uint) (Atlas, error)
	Update(id uint, a *Atlas) error
	Delete(id uint) error
	List() ([]Atlas, error)

	// Advanced query operations
	FindByName(name string) ([]Atlas, error)

	// Pagination support
	ListWithPagination(page, pageSize int) ([]Atlas, int64, error)
}
