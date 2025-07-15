package entity

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewPack(t *testing.T) {
	tests := []struct {
		name        string
		id          uuid.UUID
		size        int
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid pack creation",
			id:          uuid.New(),
			size:        250,
			expectError: false,
		},
		{
			name:        "Valid pack with large size",
			id:          uuid.New(),
			size:        5000,
			expectError: false,
		},
		{
			name:        "Invalid pack with zero size",
			id:          uuid.New(),
			size:        0,
			expectError: true,
			expectedErr: ErrPackSize,
		},
		{
			name:        "Invalid pack with negative size",
			id:          uuid.New(),
			size:        -100,
			expectError: true,
			expectedErr: ErrPackSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pack, err := NewPack(tt.id, tt.size)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				if pack != nil {
					t.Errorf("Expected pack to be nil when error occurs, got %v", pack)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if pack == nil {
				t.Errorf("Expected pack to be created, got nil")
				return
			}

			if pack.ID() != tt.id {
				t.Errorf("Expected pack ID %s, got %s", tt.id, pack.ID())
			}

			if pack.Size() != tt.size {
				t.Errorf("Expected pack size %d, got %d", tt.size, pack.Size())
			}

			if pack.CreatedAt().IsZero() {
				t.Error("Expected CreatedAt to be set")
			}

			if pack.UpdatedAt().IsZero() {
				t.Error("Expected UpdatedAt to be set")
			}
		})
	}
}

func TestPack_Size(t *testing.T) {
	id := uuid.New()
	size := 500

	pack, err := NewPack(id, size)
	if err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	if pack.Size() != size {
		t.Errorf("Expected size %d, got %d", size, pack.Size())
	}
}

func TestPack_ChangeSize(t *testing.T) {
	id := uuid.New()
	originalSize := 250

	pack, err := NewPack(id, originalSize)
	if err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	originalUpdatedAt := pack.UpdatedAt()

	tests := []struct {
		name        string
		newSize     int
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid size change",
			newSize:     500,
			expectError: false,
		},
		{
			name:        "Valid size change to larger value",
			newSize:     1000,
			expectError: false,
		},
		{
			name:        "Valid size change to smaller value",
			newSize:     100,
			expectError: false,
		},
		{
			name:        "Invalid size change to zero",
			newSize:     0,
			expectError: true,
			expectedErr: ErrPackSize,
		},
		{
			name:        "Invalid size change to negative",
			newSize:     -50,
			expectError: true,
			expectedErr: ErrPackSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pack, _ = NewPack(id, originalSize)

			err := pack.ChangeSize(tt.newSize)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				if pack.Size() != originalSize {
					t.Errorf("Expected size to remain %d after error, got %d", originalSize, pack.Size())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if pack.Size() != tt.newSize {
				t.Errorf("Expected size to be changed to %d, got %d", tt.newSize, pack.Size())
			}

			if !pack.UpdatedAt().After(originalUpdatedAt) {
				t.Errorf("Expected UpdatedAt to be updated after size change")
			}
		})
	}
}

func TestPack_ChangeSizeUpdatesTimestamp(t *testing.T) {
	id := uuid.New()
	pack, err := NewPack(id, 250)
	if err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	originalCreatedAt := pack.CreatedAt()
	originalUpdatedAt := pack.UpdatedAt()

	err = pack.ChangeSize(500)
	if err != nil {
		t.Fatalf("Failed to change pack size: %v", err)
	}

	if !pack.CreatedAt().Equal(originalCreatedAt) {
		t.Errorf("CreatedAt should not change after size change. Original: %v, Current: %v", originalCreatedAt, pack.CreatedAt())
	}

	if pack.UpdatedAt().Before(originalUpdatedAt) {
		t.Errorf("UpdatedAt should be updated after size change. Original: %v, Current: %v", originalUpdatedAt, pack.UpdatedAt())
	}
}

func TestPack_MultipleSizeChanges(t *testing.T) {
	id := uuid.New()
	pack, err := NewPack(id, 250)
	if err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	sizes := []int{500, 1000, 750, 2000}

	for _, size := range sizes {
		err := pack.ChangeSize(size)
		if err != nil {
			t.Errorf("Failed to change size to %d: %v", size, err)
			continue
		}

		if pack.Size() != size {
			t.Errorf("Expected size %d, got %d", size, pack.Size())
		}
	}

	expectedFinalSize := sizes[len(sizes)-1]
	if pack.Size() != expectedFinalSize {
		t.Errorf("Expected final size %d, got %d", expectedFinalSize, pack.Size())
	}
}

func TestPack_BaseEntityIntegration(t *testing.T) {
	id := uuid.New()
	pack, err := NewPack(id, 250)
	if err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	if pack.ID() != id {
		t.Errorf("Expected ID %s, got %s", id, pack.ID())
	}

	if pack.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if pack.UpdatedAt().IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	originalUpdatedAt := pack.UpdatedAt()

	err = pack.ChangeSize(500)
	if err != nil {
		t.Fatalf("Failed to change pack size: %v", err)
	}

	if !pack.UpdatedAt().After(originalUpdatedAt) && !pack.UpdatedAt().Equal(originalUpdatedAt) {
		t.Errorf("Expected UpdatedAt to be updated or equal after size change")
	}
}

func TestPack_SizeImmutabilityOnError(t *testing.T) {
	id := uuid.New()
	originalSize := 250
	pack, err := NewPack(id, originalSize)
	if err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	originalUpdatedAt := pack.UpdatedAt()

	err = pack.ChangeSize(-100)
	if err == nil {
		t.Error("Expected error when changing to negative size")
	}

	if pack.Size() != originalSize {
		t.Errorf("Expected size to remain %d after failed change, got %d", originalSize, pack.Size())
	}

	if !pack.UpdatedAt().Equal(originalUpdatedAt) {
		t.Errorf("Expected UpdatedAt to remain unchanged after failed size change")
	}
}
