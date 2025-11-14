package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"backend/internal/domain"
)

// Tests for InventoryUseCase

func TestInventoryUseCase_CreateWarehouse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockInventoryRepository(ctrl)
	uc := NewInventoryUseCase(mockRepo, nil, nil, zap.NewNop())

	req := CreateWarehouseRequest{Name: "Main", Code: "WH1"}

	mockRepo.EXPECT().CreateWarehouse(gomock.Any(), gomock.Any()).Return(nil)
	warehouse, err := uc.CreateWarehouse(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "Main", warehouse.Name)

	mockRepo.EXPECT().CreateWarehouse(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
	warehouse, err = uc.CreateWarehouse(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, warehouse)
}

func TestInventoryUseCase_GetWarehouse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockInventoryRepository(ctrl)
	uc := NewInventoryUseCase(mockRepo, nil, nil, zap.NewNop())

	warehouse := &domain.Warehouse{ID: 1, Name: "Main"}

	mockRepo.EXPECT().GetWarehouse(gomock.Any(), uint(1)).Return(warehouse, nil)
	result, err := uc.GetWarehouse(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, warehouse, result)

	mockRepo.EXPECT().GetWarehouse(gomock.Any(), uint(1)).Return(nil, errors.New("not found"))
	result, err = uc.GetWarehouse(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestInventoryUseCase_UpdateStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockInventoryRepository(ctrl)
	uc := NewInventoryUseCase(mockRepo, nil, nil, zap.NewNop())

	item := &domain.InventoryItem{ID: 1, WarehouseID: 1, ProductID: 2, Quantity: 10}

	// Success path
	mockRepo.EXPECT().GetInventoryItem(gomock.Any(), uint(1), uint(2)).Return(item, nil)
	mockRepo.EXPECT().UpdateInventoryItem(gomock.Any(), gomock.Any()).Return(nil)
	mockRepo.EXPECT().CreateStockMovement(gomock.Any(), gomock.Any()).Return(nil)

	req := UpdateStockRequest{OrganizationID: 1, WarehouseID: 1, ProductID: 2, MovementType: domain.StockMovementTypeReceive, Quantity: 5}
	updated, err := uc.UpdateStock(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 15, updated.Quantity)

	// Error path - update failure
	mockRepo.EXPECT().GetInventoryItem(gomock.Any(), uint(1), uint(2)).Return(item, nil)
	mockRepo.EXPECT().UpdateInventoryItem(gomock.Any(), gomock.Any()).Return(errors.New("update fail"))

	updated, err = uc.UpdateStock(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, updated)
}
