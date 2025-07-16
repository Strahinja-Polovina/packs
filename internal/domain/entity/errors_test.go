package entity

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "ErrPackSize",
			err:         ErrPackSize,
			expectedMsg: "pack size must be greater than 0",
		},
		{
			name:        "ErrPackNotFound",
			err:         ErrPackNotFound,
			expectedMsg: "pack not found",
		},
		{
			name:        "ErrOrderNotFound",
			err:         ErrOrderNotFound,
			expectedMsg: "order not found",
		},
		{
			name:        "ErrInvalidQuantity",
			err:         ErrInvalidQuantity,
			expectedMsg: "quantity must be greater than 0",
		},
		{
			name:        "ErrEmptyOrder",
			err:         ErrEmptyOrder,
			expectedMsg: "order cannot be empty",
		},
		{
			name:        "ErrInvalidAmount",
			err:         ErrInvalidAmount,
			expectedMsg: "amount must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("Expected error %s to be defined, got nil", tt.name)
				return
			}

			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, tt.err.Error())
			}
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	errors := []error{
		ErrPackSize,
		ErrPackNotFound,
		ErrOrderNotFound,
		ErrInvalidQuantity,
		ErrEmptyOrder,
		ErrInvalidAmount,
	}

	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j && err1 == err2 {
				t.Errorf("Errors at index %d and %d are the same instance: %v", i, j, err1)
			}
		}
	}
}

func TestErrorsImplementErrorInterface(t *testing.T) {
	errors := []error{
		ErrPackSize,
		ErrPackNotFound,
		ErrOrderNotFound,
		ErrInvalidQuantity,
		ErrEmptyOrder,
		ErrInvalidAmount,
	}

	for _, err := range errors {
		if err.Error() == "" {
			t.Errorf("Error %v should return a non-empty string from Error() method", err)
		}

		var testErr = err
		if testErr == nil {
			t.Errorf("Error %v should be assignable to error interface", err)
		}
	}
}

func TestPackSizeErrorUsage(t *testing.T) {
	_, err := NewPack(uuid.New(), 0)
	if !errors.Is(err, ErrPackSize) {
		t.Errorf("Expected ErrPackSize when creating pack with size 0, got %v", err)
	}

	_, err = NewPack(uuid.New(), -1)
	if !errors.Is(err, ErrPackSize) {
		t.Errorf("Expected ErrPackSize when creating pack with negative size, got %v", err)
	}

	pack, _ := NewPack(uuid.New(), 250)
	err = pack.ChangeSize(0)
	if !errors.Is(err, ErrPackSize) {
		t.Errorf("Expected ErrPackSize when changing pack size to 0, got %v", err)
	}

	err = pack.ChangeSize(-1)
	if !errors.Is(err, ErrPackSize) {
		t.Errorf("Expected ErrPackSize when changing pack size to negative, got %v", err)
	}
}

func TestInvalidQuantityErrorUsage(t *testing.T) {
	order := NewOrder(uuid.New())
	packageSize := 250

	err := order.AddItem(packageSize, 0)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("Expected ErrInvalidQuantity when adding item with quantity 0, got %v", err)
	}

	err = order.AddItem(packageSize, -1)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("Expected ErrInvalidQuantity when adding item with negative quantity, got %v", err)
	}

	_ = order.AddItem(packageSize, 1)
	err = order.UpdateItemQuantity(packageSize, 0)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("Expected ErrInvalidQuantity when updating item quantity to 0, got %v", err)
	}

	err = order.UpdateItemQuantity(packageSize, -1)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("Expected ErrInvalidQuantity when updating item quantity to negative, got %v", err)
	}

	_, err = NewOrderItem(packageSize, 0)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("Expected ErrInvalidQuantity when creating order item with quantity 0, got %v", err)
	}

	_, err = NewOrderItem(packageSize, -1)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("Expected ErrInvalidQuantity when creating order item with negative quantity, got %v", err)
	}

	item, _ := NewOrderItem(packageSize, 1)
	err = item.SetQuantity(0)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("Expected ErrInvalidQuantity when setting order item quantity to 0, got %v", err)
	}

	err = item.SetQuantity(-1)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("Expected ErrInvalidQuantity when setting order item quantity to negative, got %v", err)
	}
}

func TestOrderNotFoundErrorUsage(t *testing.T) {
	order := NewOrder(uuid.New())
	nonExistentPackageSize := 999

	err := order.RemoveItem(nonExistentPackageSize)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("Expected ErrOrderNotFound when removing non-existent item, got %v", err)
	}

	err = order.UpdateItemQuantity(nonExistentPackageSize, 1)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("Expected ErrOrderNotFound when updating non-existent item quantity, got %v", err)
	}
}

func TestPackSizeErrorForInvalidPackageSize(t *testing.T) {
	order := NewOrder(uuid.New())

	err := order.AddItem(0, 1)
	if !errors.Is(err, ErrPackSize) {
		t.Errorf("Expected ErrPackSize when adding zero package size to order, got %v", err)
	}

	_, err = NewOrderItem(0, 1)
	if !errors.Is(err, ErrPackSize) {
		t.Errorf("Expected ErrPackSize when creating order item with zero package size, got %v", err)
	}

	err = order.AddItem(-250, 1)
	if !errors.Is(err, ErrPackSize) {
		t.Errorf("Expected ErrPackSize when adding negative package size to order, got %v", err)
	}

	_, err = NewOrderItem(-250, 1)
	if !errors.Is(err, ErrPackSize) {
		t.Errorf("Expected ErrPackSize when creating order item with negative package size, got %v", err)
	}
}
