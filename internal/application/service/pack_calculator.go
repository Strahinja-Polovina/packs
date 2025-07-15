package service

import (
	"github.com/Strahinja-Polovina/packs/internal/domain/repository"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
)

// PackCalculatorService provides pack calculation functionality
type PackCalculatorService struct {
	packService  *PackService
	orderService *OrderService
}

// NewPackCalculatorService creates a new pack calculator service
func NewPackCalculatorService(packRepo repository.PackRepository, orderRepo repository.OrderRepository, logger *logger.Logger) *PackCalculatorService {
	packService := NewPackService(packRepo, logger)
	orderService := NewOrderService(orderRepo, packRepo, packService, logger)

	return &PackCalculatorService{
		packService:  packService,
		orderService: orderService,
	}
}

// GetPackService returns the underlying pack service for additional operations
func (s *PackCalculatorService) GetPackService() *PackService {
	return s.packService
}

// GetOrderService returns the underlying order service for additional operations
func (s *PackCalculatorService) GetOrderService() *OrderService {
	return s.orderService
}
