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
}