// Package main provides the entry point for the Packs API application.
//
//	@title			Packs API
//	@version		1.0
//	@description	A pack calculation and order management API
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io
//
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@host		localhost:8080
//	@BasePath	/
//
//	@schemes	http https
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Strahinja-Polovina/packs/internal/infrastructure/database"
	"github.com/Strahinja-Polovina/packs/internal/infrastructure/repository"
	"github.com/Strahinja-Polovina/packs/internal/presentation/routes"
	"github.com/Strahinja-Polovina/packs/internal/presentation/server"
	"github.com/Strahinja-Polovina/packs/pkg/config"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.SetLevel(logger.DEBUG)
	logger.Info("Starting Packs application")

	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection: %v", err)
		}
	}()

	// Initialize repositories
	packRepo := repository.NewPackPostgres(db, logger.GetLogger())
	orderRepo := repository.NewOrderPostgres(db, logger.GetLogger())

	// Create server
	srv := server.New(server.Config{
		Name:   cfg.Server.Name,
		Port:   cfg.Server.Port,
		Logger: logger.GetLogger(),
	})

	// Setup routes
	routeConfig := routes.RouteConfig{
		ServiceName: cfg.Server.Name,
		Port:        cfg.Server.Port,
		PackRepo:    packRepo,
		OrderRepo:   orderRepo,
		Logger:      logger.GetLogger(),
	}

	srv.SetupRoutes(func(router *gin.Engine) {
		routes.SetupRoutes(router, routeConfig)
	})

	// Start server in goroutine
	if err := srv.Start(); err != nil {
		logger.Fatal("Failed to start server: %v", err)
	}

	// Setup graceful shutdown
	setupGracefulShutdown(srv)

	logger.Info("Application started successfully")

	// Keep main goroutine alive and wait for shutdown
	srv.WaitForShutdown()
	logger.Info("Application shutdown complete")
}

func setupGracefulShutdown(srv *server.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Received shutdown signal")

		if err := srv.Stop(); err != nil {
			logger.Error("Error stopping server: %v", err)
		}

		// Give some time for cleanup
		time.Sleep(100 * time.Millisecond)
	}()
}
