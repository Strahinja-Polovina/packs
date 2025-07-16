package entity

import (
	"github.com/google/uuid"
)

type Order struct {
	BaseEntity
	items []OrderItem
}

type OrderItem struct {
	packageSize int
	quantity    int
}

// NewOrder creates a new order with the given ID
func NewOrder(id uuid.UUID) *Order {
	return &Order{
		BaseEntity: NewBaseEntity(id),
		items:      make([]OrderItem, 0),
	}
}

// AddItem adds a package size with quantity to the order
func (o *Order) AddItem(packageSize, quantity int) error {
	if packageSize <= 0 {
		return ErrPackSize
	}
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	for i, item := range o.items {
		if item.packageSize == packageSize {
			o.items[i].quantity += quantity
			o.Update()
			return nil
		}
	}

	o.items = append(o.items, OrderItem{
		packageSize: packageSize,
		quantity:    quantity,
	})
	o.Update()
	return nil
}

// RemoveItem removes a package size from the order
func (o *Order) RemoveItem(packageSize int) error {
	for i, item := range o.items {
		if item.packageSize == packageSize {
			o.items = append(o.items[:i], o.items[i+1:]...)
			o.Update()
			return nil
		}
	}
	return ErrOrderNotFound
}

// UpdateItemQuantity updates the quantity of a specific package size in the order
func (o *Order) UpdateItemQuantity(packageSize, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	for i, item := range o.items {
		if item.packageSize == packageSize {
			o.items[i].quantity = quantity
			o.Update()
			return nil
		}
	}
	return ErrOrderNotFound
}

// GetItems returns a copy of the order items
func (o *Order) GetItems() []OrderItem {
	items := make([]OrderItem, len(o.items))
	copy(items, o.items)
	return items
}

// GetTotalAmount calculates the total amount covered by this order
func (o *Order) GetTotalAmount() int {
	total := 0
	for _, item := range o.items {
		total += item.packageSize * item.quantity
	}
	return total
}

// IsEmpty checks if the order has no items
func (o *Order) IsEmpty() bool {
	return len(o.items) == 0
}

// Clear removes all items from the order
func (o *Order) Clear() {
	o.items = make([]OrderItem, 0)
	o.Update()
}

// NewOrderItem creates a new order item
func NewOrderItem(packageSize, quantity int) (*OrderItem, error) {
	if packageSize <= 0 {
		return nil, ErrPackSize
	}
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	return &OrderItem{
		packageSize: packageSize,
		quantity:    quantity,
	}, nil
}

// PackageSize returns the package size for this order item
func (oi *OrderItem) PackageSize() int {
	return oi.packageSize
}

// Quantity returns the quantity of this order item
func (oi *OrderItem) Quantity() int {
	return oi.quantity
}

// SetQuantity sets the quantity of this order item
func (oi *OrderItem) SetQuantity(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	oi.quantity = quantity
	return nil
}

// GetAmount returns the total amount for this order item
func (oi *OrderItem) GetAmount() int {
	return oi.packageSize * oi.quantity
}
