package service

import (
	"context"
	"fmt"

	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/Strahinja-Polovina/packs/internal/domain/repository"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/google/uuid"
)

// OrderService handles order-related business logic
type OrderService struct {
	orderRepo   repository.OrderRepository
	packRepo    repository.PackRepository
	packService *PackService
	logger      *logger.Logger
}

// NewOrderService creates a new order service
func NewOrderService(orderRepo repository.OrderRepository, packRepo repository.PackRepository, packService *PackService, logger *logger.Logger) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		packRepo:    packRepo,
		packService: packService,
		logger:      logger,
	}
}

// OrderRequest represents a request to create an order
type OrderRequest struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

// OrderResponse represents the response with order details
type OrderResponse struct {
	OrderID     uuid.UUID           `json:"order_id"`
	Amount      int                 `json:"amount"`
	PackSizes   []int               `json:"pack_sizes"`
	Combination map[int]int         `json:"combination"`
	TotalPacks  int                 `json:"total_packs"`
	TotalAmount int                 `json:"total_amount"`
	Items       []OrderItemResponse `json:"items"`
}

// OrderItemResponse represents an order item in the response
type OrderItemResponse struct {
	PackID   uuid.UUID `json:"pack_id"`
	PackSize int       `json:"pack_size"`
	Quantity int       `json:"quantity"`
	Amount   int       `json:"amount"`
}

// CreateOrderFromCalculation creates an order from pack calculation
func (s *OrderService) CreateOrderFromCalculation(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	s.logger.Info("Creating order from calculation for amount: %d", req.Amount)

	calcReq := PackCalculationRequest(req)

	calculation, err := s.packService.CalculateOptimalPacks(ctx, calcReq)
	if err != nil {
		s.logger.Error("Failed to calculate optimal packs: %v", err)
		return nil, fmt.Errorf("failed to calculate optimal packs: %w", err)
	}

	order := entity.NewOrder(uuid.New())

	var items []OrderItemResponse
	for packSize, quantity := range calculation.Combination {
		packs := s.packRepo.List(ctx)
		var pack *entity.Pack
		for _, p := range packs {
			if p.Size() == packSize {
				pack = &p
				break
			}
		}

		if pack != nil {
			err := order.AddItem(pack, quantity)
			if err != nil {
				return nil, fmt.Errorf("failed to add item to order: %w", err)
			}

			items = append(items, OrderItemResponse{
				PackID:   pack.ID(),
				PackSize: pack.Size(),
				Quantity: quantity,
				Amount:   pack.Size() * quantity,
			})
		}
	}

	err = s.orderRepo.Create(ctx, order)
	if err != nil {
		s.logger.Error("Failed to create order: %v", err)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.Info("Order created successfully with ID: %s", order.ID())

	return &OrderResponse{
		OrderID:     order.ID(),
		Amount:      calculation.Amount,
		PackSizes:   calculation.PackSizes,
		Combination: calculation.Combination,
		TotalPacks:  calculation.TotalPacks,
		TotalAmount: calculation.TotalAmount,
		Items:       items,
	}, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*OrderResponse, error) {
	s.logger.Info("Getting order with ID: %s", id)

	order, err := s.orderRepo.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get order %s: %v", id, err)
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	items := order.GetItems()
	var itemResponses []OrderItemResponse
	combination := make(map[int]int)
	totalPacks := 0
	totalAmount := 0

	for _, item := range items {
		pack := item.Pack()
		itemResponses = append(itemResponses, OrderItemResponse{
			PackID:   pack.ID(),
			PackSize: pack.Size(),
			Quantity: item.Quantity(),
			Amount:   item.GetAmount(),
		})

		combination[pack.Size()] = item.Quantity()
		totalPacks += item.Quantity()
		totalAmount += item.GetAmount()
	}

	var packSizes []int
	for size := range combination {
		packSizes = append(packSizes, size)
	}

	s.logger.Info("Order retrieved successfully with ID: %s", order.ID())

	return &OrderResponse{
		OrderID:     order.ID(),
		Amount:      totalAmount,
		PackSizes:   packSizes,
		Combination: combination,
		TotalPacks:  totalPacks,
		TotalAmount: totalAmount,
		Items:       itemResponses,
	}, nil
}

// GetAllOrders retrieves all orders
func (s *OrderService) GetAllOrders(ctx context.Context) ([]OrderResponse, error) {
	s.logger.Info("Getting all orders")

	orders := s.orderRepo.List(ctx)
	var responses []OrderResponse

	s.logger.Debug("Found %d orders to process", len(orders))

	for _, order := range orders {
		response, err := s.GetOrder(ctx, order.ID())
		if err != nil {
			s.logger.Warn("Failed to convert order %s to response: %v", order.ID(), err)
			continue
		}
		responses = append(responses, *response)
	}

	s.logger.Info("Successfully retrieved %d orders", len(responses))
	return responses, nil
}
