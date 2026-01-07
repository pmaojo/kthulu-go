// @kthulu:service:farms
package service

import (
        "car-farm-app/internal/adapters/http/modules/farms/domain"
)

type FarmsService struct {
        repo domain.FarmsRepository
}

func NewFarmsService(repo domain.FarmsRepository) domain.FarmsService {
        return &FarmsService{repo: repo}
}

func (s *FarmsService) CreateFarms(entity *domain.Farms) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *FarmsService) GetFarmsByID(id uint) (*domain.Farms, error) {
        return s.repo.GetByID(id)
}

func (s *FarmsService) UpdateFarms(entity *domain.Farms) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *FarmsService) DeleteFarms(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *FarmsService) ListFarmss(filter domain.SearchFilter) ([]*domain.Farms, error) {
        return s.repo.List(filter)
}
