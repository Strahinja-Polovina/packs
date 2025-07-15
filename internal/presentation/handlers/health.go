package handlers

import (
	"net/http"
	"time"

	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	serviceName string
	port        int
	logger      *logger.Logger
}

func NewHealthHandler(serviceName string, port int, logger *logger.Logger) *HealthHandler {
	return &HealthHandler{
		serviceName: serviceName,
		port:        port,
		logger:      logger,
	}
}

// Health handler for check application state
func (h *HealthHandler) Health(c *gin.Context) {
	h.logger.Debug("Health check requested")
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": h.serviceName,
		"port":    h.port,
		"time":    time.Now().UTC(),
	})
}
