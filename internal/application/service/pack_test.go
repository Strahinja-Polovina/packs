package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/google/uuid"
)

// MockPackRepository implements repository.PackRepository for testing
type MockPackRepository struct {
	packs []entity.Pack
}

func NewMockPackRepository() *MockPackRepository {
	pack1, _ := entity.NewPack(uuid.New(), 250)
	pack2, _ := entity.NewPack(uuid.New(), 500)
	pack3, _ := entity.NewPack(uuid.New(), 1000)
	pack4, _ := entity.NewPack(uuid.New(), 2000)
	pack5, _ := entity.NewPack(uuid.New(), 5000)

	return &MockPackRepository{
		packs: []entity.Pack{*pack1, *pack2, *pack3, *pack4, *pack5},
	}
}

func (m *MockPackRepository) List(ctx context.Context) []entity.Pack {
	return m.packs
}

func (m *MockPackRepository) Get(ctx context.Context, id uuid.UUID) (*entity.Pack, error) {
	for _, pack := range m.packs {
		if pack.ID() == id {
			return &pack, nil
		}
	}
	return nil, entity.ErrPackNotFound
}

func (m *MockPackRepository) Create(ctx context.Context, pack *entity.Pack) error {
	m.packs = append(m.packs, *pack)
	return nil
}

func (m *MockPackRepository) Update(ctx context.Context, pack *entity.Pack) error {
	for i, p := range m.packs {
		if p.ID() == pack.ID() {
			m.packs[i] = *pack
			return nil
		}
	}
	return entity.ErrPackNotFound
}

func (m *MockPackRepository) Delete(ctx context.Context, pack *entity.Pack) error {
	for i, p := range m.packs {
		if p.ID() == pack.ID() {
			m.packs = append(m.packs[:i], m.packs[i+1:]...)
			return nil
		}
	}
	return entity.ErrPackNotFound
}

func (m *MockPackRepository) ExistsBySize(ctx context.Context, size int) (bool, error) {
	for _, pack := range m.packs {
		if pack.Size() == size {
			return true, nil
		}
	}
	return false, nil
}

func TestPackService_CalculateOptimalPacks(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	tests := []struct {
		name          string
		request       PackCalculationRequest
		expectedPacks map[int]int
		expectedTotal int
		expectedWaste int
		expectError   bool
	}{
		{
			name: "Exact match with single pack size",
			request: PackCalculationRequest{
				Amount: 1000,
			},
			expectedPacks: map[int]int{1000: 1},
			expectedTotal: 1000,
			expectedWaste: 0,
			expectError:   false,
		},
		{
			name: "Multiple packs needed",
			request: PackCalculationRequest{
				Amount: 1250,
			},
			expectedPacks: map[int]int{1000: 1, 250: 1},
			expectedTotal: 1250,
			expectedWaste: 0,
			expectError:   false,
		},
		{
			name: "Amount requiring waste optimization",
			request: PackCalculationRequest{
				Amount: 1,
			},
			expectedPacks: map[int]int{250: 1},
			expectedTotal: 250,
			expectedWaste: 249,
			expectError:   false,
		},
		{
			name: "Large amount",
			request: PackCalculationRequest{
				Amount: 12001,
			},
			expectedPacks: map[int]int{5000: 2, 2000: 1, 250: 1},
			expectedTotal: 12250,
			expectedWaste: 249,
			expectError:   false,
		},
		{
			name: "Zero amount",
			request: PackCalculationRequest{
				Amount: 0,
			},
			expectError: true,
		},
		{
			name: "Negative amount",
			request: PackCalculationRequest{
				Amount: -100,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateOptimalPacks(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			if tt.expectedTotal > 0 && result.TotalAmount != tt.expectedTotal {
				t.Errorf("Expected total amount %d, got %d", tt.expectedTotal, result.TotalAmount)
			}

			if tt.expectedWaste > 0 {
				actualWaste := result.TotalAmount - tt.request.Amount
				if actualWaste != tt.expectedWaste {
					t.Errorf("Expected waste %d, got %d", tt.expectedWaste, actualWaste)
				}
			}

			if tt.expectedPacks != nil {
				if len(result.Combination) != len(tt.expectedPacks) {
					t.Errorf("Expected %d pack types, got %d", len(tt.expectedPacks), len(result.Combination))
				}

				for expectedSize, expectedCount := range tt.expectedPacks {
					if actualCount, exists := result.Combination[expectedSize]; !exists || actualCount != expectedCount {
						t.Errorf("Expected pack size %d with count %d, got count %d (exists: %v)", expectedSize, expectedCount, actualCount, exists)
					}
				}
			}

			if result.TotalAmount < tt.request.Amount {
				t.Errorf("Total amount %d is less than requested amount %d", result.TotalAmount, tt.request.Amount)
			}
		})
	}
}

func TestPackService_GetAllPacks(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	packs := service.GetAllPacks(context.Background())

	if len(packs) != 5 {
		t.Errorf("Expected 5 packs, got %d", len(packs))
	}
}

func TestPackService_CreatePack(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	newPack, err := entity.NewPack(uuid.New(), 750)
	if err != nil {
		t.Fatalf("Failed to create test pack: %v", err)
	}

	err = service.CreatePack(context.Background(), newPack)
	if err != nil {
		t.Errorf("Unexpected error creating pack: %v", err)
	}

	packs := service.GetAllPacks(context.Background())
	if len(packs) != 6 {
		t.Errorf("Expected 6 packs after creation, got %d", len(packs))
	}
}

func TestPackService_CreatePack_DuplicateSize(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	duplicatePack, err := entity.NewPack(uuid.New(), 250)
	if err != nil {
		t.Fatalf("Failed to create test pack: %v", err)
	}

	err = service.CreatePack(context.Background(), duplicatePack)
	if err == nil {
		t.Errorf("Expected error when creating pack with duplicate size, but got none")
	}

	if !errors.Is(err, entity.ErrDuplicatePackSize) {
		t.Errorf("Expected ErrDuplicatePackSize, got %v", err)
	}

	packs := service.GetAllPacks(context.Background())
	if len(packs) != 5 {
		t.Errorf("Expected 5 packs after failed creation, got %d", len(packs))
	}
}

func TestPackService_UpdatePack(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	packs := service.GetAllPacks(context.Background())
	if len(packs) == 0 {
		t.Fatal("No packs available for testing")
	}

	packToUpdate := packs[0]
	updatedPack, err := entity.NewPack(packToUpdate.ID(), 999)
	if err != nil {
		t.Fatalf("Failed to create updated pack: %v", err)
	}

	err = service.UpdatePack(context.Background(), updatedPack)
	if err != nil {
		t.Errorf("Unexpected error updating pack: %v", err)
	}

	retrievedPack, err := service.GetPackByID(context.Background(), packToUpdate.ID().String())
	if err != nil {
		t.Errorf("Failed to retrieve updated pack: %v", err)
	}

	if retrievedPack.Size() != 999 {
		t.Errorf("Expected updated pack size 999, got %d", retrievedPack.Size())
	}
}

func TestPackService_UpdatePack_DuplicateSize(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	packs := service.GetAllPacks(context.Background())
	if len(packs) < 2 {
		t.Fatal("Need at least 2 packs for testing")
	}

	packToUpdate := packs[0]
	duplicateSize := packs[1].Size()

	updatedPack, err := entity.NewPack(packToUpdate.ID(), duplicateSize)
	if err != nil {
		t.Fatalf("Failed to create updated pack: %v", err)
	}

	err = service.UpdatePack(context.Background(), updatedPack)
	if err == nil {
		t.Errorf("Expected error when updating pack to duplicate size, but got none")
	}

	if !errors.Is(err, entity.ErrDuplicatePackSize) {
		t.Errorf("Expected ErrDuplicatePackSize, got %v", err)
	}

	retrievedPack, err := service.GetPackByID(context.Background(), packToUpdate.ID().String())
	if err != nil {
		t.Errorf("Failed to retrieve pack after failed update: %v", err)
	}

	if retrievedPack.Size() == duplicateSize {
		t.Errorf("Pack size should not have been updated to duplicate size %d", duplicateSize)
	}
}

func TestPackService_UpdatePack_SameSize(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	packs := service.GetAllPacks(context.Background())
	if len(packs) == 0 {
		t.Fatal("No packs available for testing")
	}

	packToUpdate := packs[0]
	originalSize := packToUpdate.Size()

	updatedPack, err := entity.NewPack(packToUpdate.ID(), originalSize)
	if err != nil {
		t.Fatalf("Failed to create updated pack: %v", err)
	}

	err = service.UpdatePack(context.Background(), updatedPack)
	if err != nil {
		t.Errorf("Unexpected error updating pack to same size: %v", err)
	}

	retrievedPack, err := service.GetPackByID(context.Background(), packToUpdate.ID().String())
	if err != nil {
		t.Errorf("Failed to retrieve pack after update: %v", err)
	}

	if retrievedPack.Size() != originalSize {
		t.Errorf("Expected pack size %d, got %d", originalSize, retrievedPack.Size())
	}
}

func TestPackService_DeletePack(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	packs := service.GetAllPacks(context.Background())
	if len(packs) == 0 {
		t.Fatal("No packs available for testing")
	}

	packToDelete := packs[0]
	err := service.DeletePack(context.Background(), &packToDelete)
	if err != nil {
		t.Errorf("Unexpected error deleting pack: %v", err)
	}

	remainingPacks := service.GetAllPacks(context.Background())
	if len(remainingPacks) != 4 {
		t.Errorf("Expected 4 packs after deletion, got %d", len(remainingPacks))
	}

	_, err = service.GetPackByID(context.Background(), packToDelete.ID().String())
	if err == nil {
		t.Errorf("Expected error when retrieving deleted pack, but got none")
	}
}

func TestPackService_GetPackByID(t *testing.T) {
	mockRepo := NewMockPackRepository()
	service := NewPackService(mockRepo, logger.GetLogger())

	packs := service.GetAllPacks(context.Background())
	if len(packs) == 0 {
		t.Fatal("No packs available for testing")
	}

	testPack := packs[0]

	retrievedPack, err := service.GetPackByID(context.Background(), testPack.ID().String())
	if err != nil {
		t.Errorf("Unexpected error retrieving pack: %v", err)
	}

	if retrievedPack.ID() != testPack.ID() {
		t.Errorf("Expected pack ID %s, got %s", testPack.ID(), retrievedPack.ID())
	}

	_, err = service.GetPackByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Errorf("Expected error when retrieving non-existent pack, but got none")
	}

	_, err = service.GetPackByID(context.Background(), "invalid-uuid")
	if err == nil {
		t.Errorf("Expected error when retrieving pack with invalid ID, but got none")
	}
}
