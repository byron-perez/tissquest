package atlas

type RepositoryInterface interface {
    Save(a *Atlas) uint
    Retrieve(id uint) (Atlas, error)
    Update(id uint, a *Atlas) error
    Delete(id uint) error
    List() ([]Atlas, error)
}