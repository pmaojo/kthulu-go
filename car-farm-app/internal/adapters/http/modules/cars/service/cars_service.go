// @kthulu:service:cars
package service

import (
        "car-farm-app/internal/adapters/http/modules/cars/domain"
)

type CarsService struct {
        repo domain.CarsRepository
}

func NewCarsService(repo domain.CarsRepository) domain.CarsService {
        return &CarsService{repo: repo}
}

func (s *CarsService) CreateCars(entity *domain.Cars) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *CarsService) GetCarsByID(id uint) (*domain.Cars, error) {
        return s.repo.GetByID(id)
}

func (s *CarsService) UpdateCars(entity *domain.Cars) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *CarsService) DeleteCars(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *CarsService) ListCarss(filter domain.SearchFilter) ([]*domain.Cars, error) {
        return s.repo.List(filter)
}
