// @kthulu:service:inventory
package core

type inventoryService struct {
	repo InventoryRepository
}

func NewInventoryService(repo InventoryRepository) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) CreateInventory(entity *Inventory) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *inventoryService) GetInventoryByID(id uint) (*Inventory, error) {
	return s.repo.GetByID(id)
}

func (s *inventoryService) UpdateInventory(entity *Inventory) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *inventoryService) DeleteInventory(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *inventoryService) ListInventories(filter SearchFilter) ([]*Inventory, error) {
	return s.repo.List(filter)
}
