package entity

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/google/uuid"
)

func TestNewOrder(t *testing.T) {
	id := uuid.New()
	order := NewOrder(id)

	if order == nil {
		t.Fatal("Expected order to be created, got nil")
	}

	if order.ID() != id {
		t.Errorf("Expected order ID %s, got %s", id, order.ID())
	}

	if !order.IsEmpty() {
		t.Error("Expected new order to be empty")
	}

	if len(order.GetItems()) != 0 {
		t.Errorf("Expected new order to have 0 items, got %d", len(order.GetItems()))
	}

	if order.GetTotalAmount() != 0 {
		t.Errorf("Expected new order total amount to be 0, got %d", order.GetTotalAmount())
	}

	// Verify BaseEntity properties are set
	if order.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if order.UpdatedAt().IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestOrder_AddItem(t *testing.T) {
	order := NewOrder(uuid.New())
	packageSize := 250

	tests := []struct {
		name        string
		packageSize int
		quantity    int
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid item addition",
			packageSize: packageSize,
			quantity:    2,
			expectError: false,
		},
		{
			name:        "Valid item with quantity 1",
			packageSize: packageSize,
			quantity:    1,
			expectError: false,
		},
		{
			name:        "Invalid item with zero package size",
			packageSize: 0,
			quantity:    1,
			expectError: true,
			expectedErr: ErrPackSize,
		},
		{
			name:        "Invalid item with negative package size",
			packageSize: -250,
			quantity:    1,
			expectError: true,
			expectedErr: ErrPackSize,
		},
		{
			name:        "Invalid item with zero quantity",
			packageSize: packageSize,
			quantity:    0,
			expectError: true,
			expectedErr: ErrInvalidQuantity,
		},
		{
			name:        "Invalid item with negative quantity",
			packageSize: packageSize,
			quantity:    -1,
			expectError: true,
			expectedErr: ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order = NewOrder(uuid.New())

			err := order.AddItem(tt.packageSize, tt.quantity)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				if !order.IsEmpty() {
					t.Error("Expected order to remain empty after error")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if order.IsEmpty() {
				t.Error("Expected order to not be empty after adding item")
			}

			items := order.GetItems()
			if len(items) != 1 {
				t.Errorf("Expected 1 item in order, got %d", len(items))
			}

			if len(items) > 0 {
				if items[0].PackageSize() != tt.packageSize {
					t.Errorf("Expected package size %d, got %d", tt.packageSize, items[0].PackageSize())
				}
				if items[0].Quantity() != tt.quantity {
					t.Errorf("Expected quantity %d, got %d", tt.quantity, items[0].Quantity())
				}
			}
		})
	}
}

func TestOrder_AddItem_DuplicatePackageSize(t *testing.T) {
	order := NewOrder(uuid.New())
	packageSize := 250

	err := order.AddItem(packageSize, 2)
	if err != nil {
		t.Fatalf("Failed to add item first time: %v", err)
	}

	err = order.AddItem(packageSize, 3)
	if err != nil {
		t.Fatalf("Failed to add item second time: %v", err)
	}

	items := order.GetItems()
	if len(items) != 1 {
		t.Errorf("Expected 1 item in order after adding duplicate package size, got %d", len(items))
	}

	if len(items) > 0 {
		expectedQuantity := 5 // 2 + 3
		if items[0].Quantity() != expectedQuantity {
			t.Errorf("Expected quantity %d after adding duplicate package size, got %d", expectedQuantity, items[0].Quantity())
		}
	}
}

func TestOrder_RemoveItem(t *testing.T) {
	order := NewOrder(uuid.New())
	packageSize1 := 250
	packageSize2 := 500

	_ = order.AddItem(packageSize1, 2)
	_ = order.AddItem(packageSize2, 1)

	tests := []struct {
		name        string
		packageSize int
		expectError bool
		expectedErr error
	}{
		{
			name:        "Remove existing item",
			packageSize: packageSize1,
			expectError: false,
		},
		{
			name:        "Remove non-existent item",
			packageSize: 999,
			expectError: true,
			expectedErr: ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order = NewOrder(uuid.New())
			_ = order.AddItem(packageSize1, 2)
			_ = order.AddItem(packageSize2, 1)

			originalItemCount := len(order.GetItems())

			err := order.RemoveItem(tt.packageSize)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				if len(order.GetItems()) != originalItemCount {
					t.Errorf("Expected item count to remain %d after error, got %d", originalItemCount, len(order.GetItems()))
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			expectedItemCount := originalItemCount - 1
			if len(order.GetItems()) != expectedItemCount {
				t.Errorf("Expected %d items after removal, got %d", expectedItemCount, len(order.GetItems()))
			}

			items := order.GetItems()
			for _, item := range items {
				if item.PackageSize() == tt.packageSize {
					t.Errorf("Expected item with package size %d to be removed, but it still exists", tt.packageSize)
				}
			}
		})
	}
}

func TestOrder_UpdateItemQuantity(t *testing.T) {
	order := NewOrder(uuid.New())
	packageSize := 250
	_ = order.AddItem(packageSize, 2)

	tests := []struct {
		name        string
		packageSize int
		quantity    int
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid quantity update",
			packageSize: packageSize,
			quantity:    5,
			expectError: false,
		},
		{
			name:        "Update non-existent item",
			packageSize: 999,
			quantity:    3,
			expectError: true,
			expectedErr: ErrOrderNotFound,
		},
		{
			name:        "Invalid quantity - zero",
			packageSize: packageSize,
			quantity:    0,
			expectError: true,
			expectedErr: ErrInvalidQuantity,
		},
		{
			name:        "Invalid quantity - negative",
			packageSize: packageSize,
			quantity:    -1,
			expectError: true,
			expectedErr: ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order = NewOrder(uuid.New())
			_ = order.AddItem(packageSize, 2)

			err := order.UpdateItemQuantity(tt.packageSize, tt.quantity)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			items := order.GetItems()
			found := false
			for _, item := range items {
				if item.PackageSize() == tt.packageSize {
					found = true
					if item.Quantity() != tt.quantity {
						t.Errorf("Expected quantity %d, got %d", tt.quantity, item.Quantity())
					}
					break
				}
			}

			if !found {
				t.Errorf("Expected to find item with package size %d", tt.packageSize)
			}
		})
	}
}

func TestOrder_GetItems(t *testing.T) {
	order := NewOrder(uuid.New())
	packageSize1 := 250
	packageSize2 := 500

	items := order.GetItems()
	if len(items) != 0 {
		t.Errorf("Expected 0 items initially, got %d", len(items))
	}

	_ = order.AddItem(packageSize1, 2)
	_ = order.AddItem(packageSize2, 1)

	items = order.GetItems()
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	originalLength := len(order.GetItems())
	items = order.GetItems()
	_ = append(items, OrderItem{})

	if len(order.GetItems()) != originalLength {
		t.Errorf("Expected order items to remain unchanged after modifying returned slice")
	}
}

func TestOrder_GetTotalAmount(t *testing.T) {
	order := NewOrder(uuid.New())

	if order.GetTotalAmount() != 0 {
		t.Errorf("Expected total amount to be 0 initially, got %d", order.GetTotalAmount())
	}

	packageSize1 := 250
	packageSize2 := 500

	err := order.AddItem(packageSize1, 2)
	require.NoError(t, err)
	err = order.AddItem(packageSize2, 1)
	require.NoError(t, err)

	expectedTotal := 1000 // (250 * 2) + (500 * 1)
	if order.GetTotalAmount() != expectedTotal {
		t.Errorf("Expected total amount %d, got %d", expectedTotal, order.GetTotalAmount())
	}
}

func TestOrder_IsEmpty(t *testing.T) {
	order := NewOrder(uuid.New())

	if !order.IsEmpty() {
		t.Error("Expected new order to be empty")
	}

	packageSize := 250
	_ = order.AddItem(packageSize, 1)

	if order.IsEmpty() {
		t.Error("Expected order to not be empty after adding item")
	}

	_ = order.RemoveItem(packageSize)

	if !order.IsEmpty() {
		t.Error("Expected order to be empty after removing all items")
	}
}

func TestOrder_Clear(t *testing.T) {
	order := NewOrder(uuid.New())
	packageSize1 := 250
	packageSize2 := 500

	_ = order.AddItem(packageSize1, 2)
	_ = order.AddItem(packageSize2, 1)

	if order.IsEmpty() {
		t.Error("Expected order to have items before clearing")
	}

	originalUpdatedAt := order.UpdatedAt()
	order.Clear()

	if !order.IsEmpty() {
		t.Error("Expected order to be empty after clearing")
	}

	if len(order.GetItems()) != 0 {
		t.Errorf("Expected 0 items after clearing, got %d", len(order.GetItems()))
	}

	if order.GetTotalAmount() != 0 {
		t.Errorf("Expected total amount to be 0 after clearing, got %d", order.GetTotalAmount())
	}

	if !order.UpdatedAt().After(originalUpdatedAt) && !order.UpdatedAt().Equal(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after clearing")
	}
}

func TestNewOrderItem(t *testing.T) {
	packageSize := 250

	tests := []struct {
		name        string
		packageSize int
		quantity    int
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid order item creation",
			packageSize: packageSize,
			quantity:    2,
			expectError: false,
		},
		{
			name:        "Invalid order item with zero package size",
			packageSize: 0,
			quantity:    1,
			expectError: true,
			expectedErr: ErrPackSize,
		},
		{
			name:        "Invalid order item with negative package size",
			packageSize: -250,
			quantity:    1,
			expectError: true,
			expectedErr: ErrPackSize,
		},
		{
			name:        "Invalid order item with zero quantity",
			packageSize: packageSize,
			quantity:    0,
			expectError: true,
			expectedErr: ErrInvalidQuantity,
		},
		{
			name:        "Invalid order item with negative quantity",
			packageSize: packageSize,
			quantity:    -1,
			expectError: true,
			expectedErr: ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewOrderItem(tt.packageSize, tt.quantity)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				if item != nil {
					t.Errorf("Expected item to be nil when error occurs, got %v", item)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if item == nil {
				t.Error("Expected item to be created, got nil")
				return
			}

			if item.PackageSize() != tt.packageSize {
				t.Errorf("Expected package size %d, got %d", tt.packageSize, item.PackageSize())
			}

			if item.Quantity() != tt.quantity {
				t.Errorf("Expected quantity %d, got %d", tt.quantity, item.Quantity())
			}
		})
	}
}

func TestOrderItem_PackageSize(t *testing.T) {
	packageSize := 250
	item, _ := NewOrderItem(packageSize, 2)

	if item.PackageSize() != packageSize {
		t.Errorf("Expected package size %d, got %d", packageSize, item.PackageSize())
	}
}

func TestOrderItem_Quantity(t *testing.T) {
	packageSize := 250
	quantity := 3
	item, _ := NewOrderItem(packageSize, quantity)

	if item.Quantity() != quantity {
		t.Errorf("Expected quantity %d, got %d", quantity, item.Quantity())
	}
}

func TestOrderItem_SetQuantity(t *testing.T) {
	packageSize := 250
	item, _ := NewOrderItem(packageSize, 2)

	tests := []struct {
		name        string
		quantity    int
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid quantity update",
			quantity:    5,
			expectError: false,
		},
		{
			name:        "Invalid quantity - zero",
			quantity:    0,
			expectError: true,
			expectedErr: ErrInvalidQuantity,
		},
		{
			name:        "Invalid quantity - negative",
			quantity:    -1,
			expectError: true,
			expectedErr: ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, _ = NewOrderItem(packageSize, 2)
			originalQuantity := item.Quantity()

			err := item.SetQuantity(tt.quantity)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				if item.Quantity() != originalQuantity {
					t.Errorf("Expected quantity to remain %d after error, got %d", originalQuantity, item.Quantity())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if item.Quantity() != tt.quantity {
				t.Errorf("Expected quantity %d, got %d", tt.quantity, item.Quantity())
			}
		})
	}
}

func TestOrderItem_GetAmount(t *testing.T) {
	packageSize := 250
	quantity := 3
	item, _ := NewOrderItem(packageSize, quantity)

	expectedAmount := 750 // 250 * 3
	if item.GetAmount() != expectedAmount {
		t.Errorf("Expected amount %d, got %d", expectedAmount, item.GetAmount())
	}
}

func TestOrder_Integration(t *testing.T) {
	order := NewOrder(uuid.New())
	packageSize1 := 250
	packageSize2 := 500
	packageSize3 := 1000

	_ = order.AddItem(packageSize1, 2)
	_ = order.AddItem(packageSize2, 1)
	_ = order.AddItem(packageSize3, 1)

	expectedTotal := 2000 // (250 * 2) + (500 * 1) + (1000 * 1)
	if order.GetTotalAmount() != expectedTotal {
		t.Errorf("Expected total amount %d, got %d", expectedTotal, order.GetTotalAmount())
	}

	_ = order.UpdateItemQuantity(packageSize1, 4)
	expectedTotal = 2500 // (250 * 4) + (500 * 1) + (1000 * 1)
	if order.GetTotalAmount() != expectedTotal {
		t.Errorf("Expected total amount after update %d, got %d", expectedTotal, order.GetTotalAmount())
	}

	_ = order.RemoveItem(packageSize2)
	expectedTotal = 2000 // (250 * 4) + (1000 * 1)
	if order.GetTotalAmount() != expectedTotal {
		t.Errorf("Expected total amount after removal %d, got %d", expectedTotal, order.GetTotalAmount())
	}

	if len(order.GetItems()) != 2 {
		t.Errorf("Expected 2 items after removal, got %d", len(order.GetItems()))
	}

	order.Clear()
	if !order.IsEmpty() {
		t.Error("Expected order to be empty after clearing")
	}
	if order.GetTotalAmount() != 0 {
		t.Errorf("Expected total amount to be 0 after clearing, got %d", order.GetTotalAmount())
	}
}
