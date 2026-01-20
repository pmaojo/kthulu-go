// @kthulu:service:product
package core

type productService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(entity *Product) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *productService) GetProductByID(id uint) (*Product, error) {
	return s.repo.GetByID(id)
}

func (s *productService) UpdateProduct(entity *Product) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *productService) DeleteProduct(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *productService) ListProducts(filter SearchFilter) ([]*Product, error) {
	return s.repo.List(filter)
}
