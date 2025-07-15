package routes

import (
	"github.com/Strahinja-Polovina/packs/internal/application/service"
	"github.com/Strahinja-Polovina/packs/internal/domain/repository"
	"github.com/Strahinja-Polovina/packs/internal/presentation/handlers"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouteConfig struct {
	ServiceName string
	Port        int
	PackRepo    repository.PackRepository
	OrderRepo   repository.OrderRepository
	Logger      *logger.Logger
}

func SetupRoutes(router *gin.Engine, config RouteConfig) {
	// Initialize services
	packCalculatorService := service.NewPackCalculatorService(config.PackRepo, config.OrderRepo, config.Logger)
	orderService := packCalculatorService.GetOrderService()
	packService := service.NewPackService(config.PackRepo, config.Logger)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(config.ServiceName, config.Port, config.Logger)
	packCalculatorHandler := handlers.NewPackCalculatorHandler(packCalculatorService, config.Logger)
	orderHandler := handlers.NewOrderHandler(orderService, config.Logger)
	webHandler := handlers.NewWebHandler(packService, orderService, config.Logger)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check route
	router.GET("/health", healthHandler.Health)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Pack-sizes CRUD routes
		v1.GET("/pack-sizes", packCalculatorHandler.GetPackSizes)
		v1.POST("/pack-sizes", packCalculatorHandler.CreatePackSize)
		v1.PUT("/pack-sizes/:id", packCalculatorHandler.UpdatePackSize)
		v1.DELETE("/pack-sizes/:id", packCalculatorHandler.DeletePackSize)

		// Order routes
		v1.POST("/orders", orderHandler.CreateOrder)
		v1.GET("/orders", orderHandler.GetAllOrders)
	}

	// Web routes
	web := router.Group("/web")
	{
		// Package management routes
		web.GET("/packages/new", webHandler.GetPackageForm)
		web.GET("/packages/:id/edit", webHandler.GetPackageEditForm)
		web.GET("/packages/table", webHandler.GetPackagesTableBody)
		web.POST("/packages", webHandler.HandlePackageCreation)
		web.PUT("/packages/:id", webHandler.HandlePackageUpdate)
		web.DELETE("/packages/:id", webHandler.HandlePackageDelete)

		// Order management routes
		web.GET("/orders", webHandler.GetOrdersList)
		web.POST("/orders", webHandler.HandleOrderCreation)
	}

	// Main page route
	router.GET("/", webHandler.Index)
}
