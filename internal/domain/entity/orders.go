package entity

import (
	"github.com/google/uuid"
)

type Order struct {
	BaseEntity
	items []OrderItem
}

type OrderItem struct {
	pack     *Pack
	quantity int
}

// NewOrder creates a new order with the given ID
func NewOrder(id uuid.UUID) *Order {
	return &Order{
		BaseEntity: NewBaseEntity(id),
		items:      make([]OrderItem, 0),
	}
}

// AddItem adds a pack with quantity to the order
func (o *Order) AddItem(pack *Pack, quantity int) error {
	if pack == nil {
		return ErrPackSize
	}
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	for i, item := range o.items {
		if item.pack.ID() == pack.ID() {
			o.items[i].quantity += quantity
			o.Update()
			return nil
		}
	}

	o.items = append(o.items, OrderItem{
		pack:     pack,
		quantity: quantity,
	})
	o.Update()
	return nil
}

// RemoveItem removes a pack from the order
func (o *Order) RemoveItem(packID uuid.UUID) error {
	for i, item := range o.items {
		if item.pack.ID() == packID {
			o.items = append(o.items[:i], o.items[i+1:]...)
			o.Update()
			return nil
		}
	}
	return ErrOrderNotFound
}

// UpdateItemQuantity updates the quantity of a specific pack in the order
func (o *Order) UpdateItemQuantity(packID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	for i, item := range o.items {
		if item.pack.ID() == packID {
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
		total += item.pack.Size() * item.quantity
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
func NewOrderItem(pack *Pack, quantity int) (*OrderItem, error) {
	if pack == nil {
		return nil, ErrPackSize
	}
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	return &OrderItem{
		pack:     pack,
		quantity: quantity,
	}, nil
}

// Pack returns the pack associated with this order item
func (oi *OrderItem) Pack() *Pack {
	return oi.pack
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
	return oi.pack.Size() * oi.quantity
}
