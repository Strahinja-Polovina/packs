package handlers

import (
	"net/http"

	"github.com/Strahinja-Polovina/packs/internal/application/service"
	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/Strahinja-Polovina/packs/internal/presentation/templates"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WebHandler handles web requests for the frontend
type WebHandler struct {
	packService  *service.PackService
	orderService *service.OrderService
	logger       *logger.Logger
}

// NewWebHandler creates a new web handler
func NewWebHandler(packService *service.PackService, orderService *service.OrderService, logger *logger.Logger) *WebHandler {
	return &WebHandler{
		packService:  packService,
		orderService: orderService,
		logger:       logger,
	}
}

// Index serves the main page
func (h *WebHandler) Index(c *gin.Context) {
	h.logger.Info("Serving main page")

	packs := h.packService.GetAllPacks(c.Request.Context())

	orders, err := h.orderService.GetAllOrders(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get orders: %v", err)
		orders = []service.OrderResponse{}
	}

	component := templates.Index(packs, orders)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.Error("Failed to render index template: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Template rendering failed",
			Message: err.Error(),
		})
		return
	}
}

// GetPackageForm serves the package creation form
func (h *WebHandler) GetPackageForm(c *gin.Context) {
	h.logger.Info("Serving package creation form")

	component := templates.PackageForm(nil, false)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.Error("Failed to render package form template: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Template rendering failed",
			Message: err.Error(),
		})
		return
	}
}

// GetPackageEditForm serves the package edit form
func (h *WebHandler) GetPackageEditForm(c *gin.Context) {
	id := c.Param("id")
	h.logger.Info("Serving package edit form for ID: %s", id)

	pack, err := h.packService.GetPackByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get pack: %v", err)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Pack not found",
			Message: err.Error(),
		})
		return
	}

	component := templates.PackageForm(pack, true)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.Error("Failed to render package edit form template: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Template rendering failed",
			Message: err.Error(),
		})
		return
	}
}

// GetPackagesTableBody serves the packages table body for HTMX updates
func (h *WebHandler) GetPackagesTableBody(c *gin.Context) {
	h.logger.Info("Serving packages table body")

	packs := h.packService.GetAllPacks(c.Request.Context())

	c.Header("Content-Type", "text/html")
	for _, pack := range packs {
		component := templates.PackageRow(pack)
		if err := component.Render(c.Request.Context(), c.Writer); err != nil {
			h.logger.Error("Failed to render package row template: %v", err)
			continue
		}
	}
}

// GetOrdersList serves the orders list for HTMX updates
func (h *WebHandler) GetOrdersList(c *gin.Context) {
	h.logger.Info("Serving orders list")

	orders, err := h.orderService.GetAllOrders(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get orders: %v", err)
		orders = []service.OrderResponse{}
	}

	c.Header("Content-Type", "text/html")
	for _, order := range orders {
		component := templates.OrderCard(order)
		if err := component.Render(c.Request.Context(), c.Writer); err != nil {
			h.logger.Error("Failed to render order card template: %v", err)
			continue
		}
	}
}

// HandleOrderCreation handles order creation and returns the result
func (h *WebHandler) HandleOrderCreation(c *gin.Context) {
	h.logger.Info("Handling order creation from web")

	var req service.OrderRequest
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("Invalid request format: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	result, err := h.orderService.CreateOrderFromCalculation(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Order creation failed: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Order creation failed",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Order created successfully with ID: %s", result.OrderID)

	component := templates.OrderResult(*result)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.Error("Failed to render order result template: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Template rendering failed",
			Message: err.Error(),
		})
		return
	}
}

// HandlePackageCreation handles package creation and returns updated table
func (h *WebHandler) HandlePackageCreation(c *gin.Context) {
	h.logger.Info("Handling package creation from web")

	var req struct {
		Size int `form:"size" json:"size" binding:"required,min=1"`
	}

	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("Invalid request format: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	pack, err := entity.NewPack(uuid.New(), req.Size)
	if err != nil {
		h.logger.Error("Failed to create pack entity: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid pack data",
			Message: err.Error(),
		})
		return
	}

	err = h.packService.CreatePack(c.Request.Context(), pack)
	if err != nil {
		h.logger.Error("Failed to create pack: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create pack",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Pack created successfully with ID: %s", pack.ID())

	h.GetPackagesTableBody(c)
}

// HandlePackageUpdate handles package update and returns updated table
func (h *WebHandler) HandlePackageUpdate(c *gin.Context) {
	id := c.Param("id")
	h.logger.Info("Handling package update for ID: %s", id)

	var req struct {
		Size int `form:"size" json:"size" binding:"required,min=1"`
	}

	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("Invalid request format: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	pack, err := h.packService.GetPackByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get pack: %v", err)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Pack not found",
			Message: err.Error(),
		})
		return
	}

	err = pack.ChangeSize(req.Size)
	if err != nil {
		h.logger.Error("Failed to change pack size: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid pack size",
			Message: err.Error(),
		})
		return
	}

	err = h.packService.UpdatePack(c.Request.Context(), pack)
	if err != nil {
		h.logger.Error("Failed to update pack: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update pack",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Pack updated successfully with ID: %s", pack.ID())

	h.GetPackagesTableBody(c)
}

// HandlePackageDelete handles package deletion and returns updated table
func (h *WebHandler) HandlePackageDelete(c *gin.Context) {
	id := c.Param("id")
	h.logger.Info("Handling package deletion for ID: %s", id)

	pack, err := h.packService.GetPackByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get pack: %v", err)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Pack not found",
			Message: err.Error(),
		})
		return
	}

	err = h.packService.DeletePack(c.Request.Context(), pack)
	if err != nil {
		h.logger.Error("Failed to delete pack: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete pack",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Pack deleted successfully with ID: %s", pack.ID())

	h.GetPackagesTableBody(c)
}
