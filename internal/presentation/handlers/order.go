package handlers

import (
	"net/http"

	"github.com/Strahinja-Polovina/packs/internal/application/service"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	service *service.OrderService
	logger  *logger.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *service.OrderService, logger *logger.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

// CreateOrder handles POST /api/v1/orders
// @Summary Create a new order
// @Description Create a new order from pack calculation
// @Tags orders
// @Accept json
// @Produce json
// @Param request body service.OrderRequest true "Order creation request"
// @Success 201 {object} service.OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	h.logger.Info("Received create order request")

	var req service.OrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request format: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	result, err := h.service.CreateOrderFromCalculation(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Order creation failed: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Order creation failed",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Order created successfully with ID: %s", result.OrderID)
	c.JSON(http.StatusCreated, result)
}

// GetAllOrders handles GET /api/v1/orders
// @Summary Get all orders
// @Description Retrieve all orders from the system
// @Tags orders
// @Produce json
// @Success 200 {array} service.OrderResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [get]
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	h.logger.Info("Received get all orders request")

	orders, err := h.service.GetAllOrders(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to retrieve orders: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve orders",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Successfully retrieved %d orders", len(orders))
	c.JSON(http.StatusOK, orders)
}
