// @kthulu:service:maintenance
package service

import (
        "car-farm-app/internal/adapters/http/modules/maintenance/domain"
)

type MaintenanceService struct {
        repo domain.MaintenanceRepository
}

func NewMaintenanceService(repo domain.MaintenanceRepository) domain.MaintenanceService {
        return &MaintenanceService{repo: repo}
}

func (s *MaintenanceService) CreateMaintenance(entity *domain.Maintenance) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *MaintenanceService) GetMaintenanceByID(id uint) (*domain.Maintenance, error) {
        return s.repo.GetByID(id)
}

func (s *MaintenanceService) UpdateMaintenance(entity *domain.Maintenance) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *MaintenanceService) DeleteMaintenance(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *MaintenanceService) ListMaintenances(filter domain.SearchFilter) ([]*domain.Maintenance, error) {
        return s.repo.List(filter)
}
