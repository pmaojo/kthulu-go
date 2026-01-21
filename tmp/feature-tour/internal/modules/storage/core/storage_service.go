// @kthulu:service:storage
package core

type storageService struct {
	repo StorageRepository
}

func NewStorageService(repo StorageRepository) StorageService {
	return &storageService{repo: repo}
}

func (s *storageService) CreateStorage(entity *Storage) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *storageService) GetStorageByID(id uint) (*Storage, error) {
	return s.repo.GetByID(id)
}

func (s *storageService) UpdateStorage(entity *Storage) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *storageService) DeleteStorage(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *storageService) ListStorages(filter SearchFilter) ([]*Storage, error) {
	return s.repo.List(filter)
}
