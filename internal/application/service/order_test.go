package service

import (
	"context"
	"testing"

	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/google/uuid"
)

// MockOrderRepository implements repository.OrderRepository for testing
type MockOrderRepository struct {
	orders []entity.Order
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders: []entity.Order{},
	}
}

func (m *MockOrderRepository) List(ctx context.Context) []entity.Order {
	return m.orders
}

func (m *MockOrderRepository) Get(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	for _, order := range m.orders {
		if order.ID() == id {
			return &order, nil
		}
	}
	return nil, entity.ErrOrderNotFound
}

func (m *MockOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	m.orders = append(m.orders, *order)
	return nil
}

func (m *MockOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	for i, o := range m.orders {
		if o.ID() == order.ID() {
			m.orders[i] = *order
			return nil
		}
	}
	return entity.ErrOrderNotFound
}

func (m *MockOrderRepository) Delete(ctx context.Context, order *entity.Order) error {
	for i, o := range m.orders {
		if o.ID() == order.ID() {
			m.orders = append(m.orders[:i], m.orders[i+1:]...)
			return nil
		}
	}
	return entity.ErrOrderNotFound
}

func TestOrderService_CreateOrderFromCalculation(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository()
	mockPackRepo := NewMockPackRepository()
	packService := NewPackService(mockPackRepo, logger.GetLogger())
	orderService := NewOrderService(mockOrderRepo, mockPackRepo, packService, logger.GetLogger())

	tests := []struct {
		name        string
		request     OrderRequest
		expectError bool
	}{
		{
			name: "Valid order creation",
			request: OrderRequest{
				Amount: 1250,
			},
			expectError: false,
		},
		{
			name: "Invalid amount - zero",
			request: OrderRequest{
				Amount: 0,
			},
			expectError: true,
		},
		{
			name: "Invalid amount - negative",
			request: OrderRequest{
				Amount: -100,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := orderService.CreateOrderFromCalculation(context.Background(), tt.request)

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

			if result.Amount != tt.request.Amount {
				t.Errorf("Expected order amount %d, got %d", tt.request.Amount, result.Amount)
			}

			if len(result.Items) == 0 {
				t.Errorf("Expected order to have items, but got none")
			}

			if result.TotalAmount < tt.request.Amount {
				t.Errorf("Total amount %d is less than requested amount %d", result.TotalAmount, tt.request.Amount)
			}

			orders, err := orderService.GetAllOrders(context.Background())
			if err != nil {
				t.Errorf("Unexpected error getting all orders: %v", err)
			}
			if len(orders) == 0 {
				t.Errorf("Expected order to be stored, but repository is empty")
			}
		})
	}
}

func TestOrderService_GetOrder(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository()
	mockPackRepo := NewMockPackRepository()
	packService := NewPackService(mockPackRepo, logger.GetLogger())
	orderService := NewOrderService(mockOrderRepo, mockPackRepo, packService, logger.GetLogger())

	orderRequest := OrderRequest{Amount: 1000}
	createdOrder, err := orderService.CreateOrderFromCalculation(context.Background(), orderRequest)
	if err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}

	tests := []struct {
		name        string
		orderID     uuid.UUID
		expectError bool
	}{
		{
			name:        "Get existing order",
			orderID:     createdOrder.OrderID,
			expectError: false,
		},
		{
			name:        "Get non-existent order",
			orderID:     uuid.New(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := orderService.GetOrder(context.Background(), tt.orderID)

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

			if result.OrderID != tt.orderID {
				t.Errorf("Expected order ID %s, got %s", tt.orderID, result.OrderID)
			}
		})
	}
}

func TestOrderService_GetAllOrders(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository()
	mockPackRepo := NewMockPackRepository()
	packService := NewPackService(mockPackRepo, logger.GetLogger())
	orderService := NewOrderService(mockOrderRepo, mockPackRepo, packService, logger.GetLogger())

	orders, err := orderService.GetAllOrders(context.Background())
	if err != nil {
		t.Errorf("Unexpected error getting all orders: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("Expected 0 orders initially, got %d", len(orders))
	}

	orderRequests := []OrderRequest{
		{Amount: 1000},
		{Amount: 2500},
		{Amount: 750},
	}

	for _, req := range orderRequests {
		_, err := orderService.CreateOrderFromCalculation(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to create test order: %v", err)
		}
	}

	orders, err = orderService.GetAllOrders(context.Background())
	if err != nil {
		t.Errorf("Unexpected error getting all orders: %v", err)
	}
	if len(orders) != 3 {
		t.Errorf("Expected 3 orders, got %d", len(orders))
	}
}

func TestOrderService_Integration(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository()
	mockPackRepo := NewMockPackRepository()
	packService := NewPackService(mockPackRepo, logger.GetLogger())
	orderService := NewOrderService(mockOrderRepo, mockPackRepo, packService, logger.GetLogger())

	orderRequest := OrderRequest{Amount: 1250}
	createdOrder, err := orderService.CreateOrderFromCalculation(context.Background(), orderRequest)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	if createdOrder.TotalAmount != 1250 {
		t.Errorf("Expected total amount 1250, got %d", createdOrder.TotalAmount)
	}

	expectedWaste := 0
	actualWaste := createdOrder.TotalAmount - orderRequest.Amount
	if actualWaste != expectedWaste {
		t.Errorf("Expected waste %d, got %d", expectedWaste, actualWaste)
	}

	if len(createdOrder.Items) != 2 {
		t.Errorf("Expected 2 order items, got %d", len(createdOrder.Items))
	}

	expectedItems := map[int]int{1000: 1, 250: 1}
	for _, item := range createdOrder.Items {
		if expectedCount, exists := expectedItems[item.PackSize]; !exists || item.Quantity != expectedCount {
			t.Errorf("Unexpected item: pack size %d, quantity %d", item.PackSize, item.Quantity)
		}
	}
}
