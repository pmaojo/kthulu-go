// @kthulu:service:cache
package core

type cacheService struct {
	repo CacheRepository
}

func NewCacheService(repo CacheRepository) CacheService {
	return &cacheService{repo: repo}
}

func (s *cacheService) CreateCache(entity *Cache) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *cacheService) GetCacheByID(id uint) (*Cache, error) {
	return s.repo.GetByID(id)
}

func (s *cacheService) UpdateCache(entity *Cache) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *cacheService) DeleteCache(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *cacheService) ListCaches(filter SearchFilter) ([]*Cache, error) {
	return s.repo.List(filter)
}
