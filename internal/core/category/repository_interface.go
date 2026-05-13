package category

type RepositoryInterface interface {
	Save(c *Category) uint
	Retrieve(id uint) (*Category, error)
	Update(id uint, c *Category) error
	Delete(id uint) error
	List() ([]Category, error)

	FindByType(categoryType CategoryType) ([]Category, error)
	FindByParent(parentID uint) ([]Category, error)
	FindRootCategories() ([]Category, error)
	FindByMetacategory(metacategoryID uint) ([]Category, error)
	// ListWithCounts returns all categories with the number of tissue records
	// directly associated with each one (not including descendants).
	ListWithCounts() ([]CategoryWithCount, error)
}

// CategoryWithCount is a read model for the filter panel.
type CategoryWithCount struct {
	Category
	Count int
}
