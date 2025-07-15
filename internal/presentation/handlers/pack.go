package handlers

import (
	"errors"
	"net/http"

	"github.com/Strahinja-Polovina/packs/internal/application/service"
	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PackCalculatorHandler handles HTTP requests for pack calculations
type PackCalculatorHandler struct {
	service *service.PackCalculatorService
	logger  *logger.Logger
}

// NewPackCalculatorHandler creates a new pack calculator handler
func NewPackCalculatorHandler(service *service.PackCalculatorService, logger *logger.Logger) *PackCalculatorHandler {
	return &PackCalculatorHandler{
		service: service,
		logger:  logger,
	}
}

// GetPackSizes handles GET /api/v1/pack-sizes
// @Summary Get available pack sizes
// @Description Get all available pack sizes from the system
// @Tags packs
// @Produce json
// @Success 200 {object} PackSizesResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pack-sizes [get]
func (h *PackCalculatorHandler) GetPackSizes(c *gin.Context) {
	h.logger.Info("Received get pack sizes request")

	packs := h.service.GetPackService().GetAllPacks(c.Request.Context())

	// Convert pack entities to PackResponse objects
	packResponses := make([]PackResponse, len(packs))
	for i, pack := range packs {
		packResponses[i] = PackResponse{
			ID:   pack.ID(),
			Size: pack.Size(),
		}
	}

	h.logger.Info("Successfully retrieved %d pack sizes", len(packResponses))
	c.JSON(http.StatusOK, PackSizesResponse{
		Packs: packResponses,
		Count: len(packResponses),
	})
}

// CreatePackSize handles POST /api/v1/pack-sizes
// @Summary Create a new pack size
// @Description Add a new pack size to the system
// @Tags packs
// @Accept json
// @Produce json
// @Param request body CreatePackSizeRequest true "Pack size creation request"
// @Success 201 {object} PackResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pack-sizes [post]
func (h *PackCalculatorHandler) CreatePackSize(c *gin.Context) {
	h.logger.Info("Received create pack size request")

	var req CreatePackSizeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request format for create pack size: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	pack, err := entity.NewPack(uuid.New(), req.Size)
	if err != nil {
		h.logger.Error("Invalid pack size %d: %v", req.Size, err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid pack size",
			Message: err.Error(),
		})
		return
	}

	err = h.service.GetPackService().CreatePack(c.Request.Context(), pack)
	if err != nil {
		if errors.Is(err, entity.ErrDuplicatePackSize) {
			h.logger.Warn("Attempted to create duplicate pack size: %d", req.Size)
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Duplicate pack size",
				Message: "A pack with this size already exists",
			})
		} else {
			h.logger.Error("Failed to create pack size %d: %v", req.Size, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to create pack size",
				Message: err.Error(),
			})
		}
		return
	}

	h.logger.Info("Pack size created successfully with ID: %s, size: %d", pack.ID(), pack.Size())
	c.JSON(http.StatusCreated, PackResponse{
		ID:   pack.ID(),
		Size: pack.Size(),
	})
}

// UpdatePackSize handles PUT /api/v1/pack-sizes/:id
// @Summary Update a pack size
// @Description Update an existing pack size
// @Tags packs
// @Accept json
// @Produce json
// @Param id path string true "Pack ID" format(uuid)
// @Param request body UpdatePackSizeRequest true "Pack size update request"
// @Success 200 {object} PackResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pack-sizes/{id} [put]
func (h *PackCalculatorHandler) UpdatePackSize(c *gin.Context) {
	idStr := c.Param("id")
	h.logger.Info("Received update pack size request for ID: %s", idStr)

	packID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid pack ID format: %s", idStr)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid pack ID",
			Message: "Pack ID must be a valid UUID",
		})
		return
	}

	var req UpdatePackSizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request format for update pack size: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	pack, err := h.service.GetPackService().GetPackByID(c.Request.Context(), packID.String())
	if err != nil {
		h.logger.Error("Pack not found for update with ID: %s", packID)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Pack not found",
			Message: err.Error(),
		})
		return
	}

	err = pack.ChangeSize(req.Size)
	if err != nil {
		h.logger.Error("Invalid pack size %d for pack %s: %v", req.Size, packID, err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid pack size",
			Message: err.Error(),
		})
		return
	}

	err = h.service.GetPackService().UpdatePack(c.Request.Context(), pack)
	if err != nil {
		if errors.Is(err, entity.ErrDuplicatePackSize) {
			h.logger.Warn("Attempted to update pack %s to duplicate size: %d", packID, req.Size)
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Duplicate pack size",
				Message: "A pack with this size already exists",
			})
		} else {
			h.logger.Error("Failed to update pack %s: %v", packID, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update pack size",
				Message: err.Error(),
			})
		}
		return
	}

	h.logger.Info("Pack updated successfully with ID: %s, new size: %d", packID, pack.Size())
	c.JSON(http.StatusOK, PackResponse{
		ID:   pack.ID(),
		Size: pack.Size(),
	})
}

// DeletePackSize handles DELETE /api/v1/pack-sizes/:id
// @Summary Delete a pack size
// @Description Remove a pack size from the system
// @Tags packs
// @Param id path string true "Pack ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pack-sizes/{id} [delete]
func (h *PackCalculatorHandler) DeletePackSize(c *gin.Context) {
	idStr := c.Param("id")
	h.logger.Info("Received delete pack size request for ID: %s", idStr)

	packID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid pack ID format for deletion: %s", idStr)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid pack ID",
			Message: "Pack ID must be a valid UUID",
		})
		return
	}

	pack, err := h.service.GetPackService().GetPackByID(c.Request.Context(), packID.String())
	if err != nil {
		h.logger.Error("Pack not found for deletion with ID: %s", packID)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Pack not found",
			Message: err.Error(),
		})
		return
	}

	err = h.service.GetPackService().DeletePack(c.Request.Context(), pack)
	if err != nil {
		h.logger.Error("Failed to delete pack %s: %v", packID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete pack size",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Pack deleted successfully with ID: %s", packID)
	c.Status(http.StatusNoContent)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// PackSizesResponse represents the response for pack sizes endpoint
type PackSizesResponse struct {
	Packs []PackResponse `json:"packs"`
	Count int            `json:"count"`
}

// CreatePackSizeRequest represents a request to create a pack size
type CreatePackSizeRequest struct {
	Size int `json:"size" binding:"required,min=1"`
}

// UpdatePackSizeRequest represents a request to update a pack size
type UpdatePackSizeRequest struct {
	Size int `json:"size" binding:"required,min=1"`
}

// PackResponse represents a pack in the response
type PackResponse struct {
	ID   uuid.UUID `json:"id"`
	Size int       `json:"size"`
}
