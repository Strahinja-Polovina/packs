package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"

	"github.com/Strahinja-Polovina/packs/pkg/logger"
)

type Server struct {
	name       string
	port       int
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	logger     *logger.Logger
	httpServer *http.Server
	router     *gin.Engine
}

type Config struct {
	Name   string
	Port   int
	Logger *logger.Logger
}

func New(config Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	serverLogger := config.Logger
	if serverLogger == nil {
		serverLogger = logger.GetLogger()
	}

	// Set gin mode to release for production
	gin.SetMode(gin.ReleaseMode)

	// Create gin router
	router := gin.New()
	router.Use(gin.Recovery())

	server := &Server{
		name:   config.Name,
		port:   config.Port,
		ctx:    ctx,
		cancel: cancel,
		logger: serverLogger,
		router: router,
	}

	return server
}

func (s *Server) SetupRoutes(setupFunc func(*gin.Engine)) {
	setupFunc(s.router)
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server %s is already running", s.name)
	}

	s.logger.Info("Starting server %s on port %d", s.name, s.port)
	s.running = true

	// Start server in a separate goroutine
	s.wg.Add(1)
	go s.run()

	s.logger.Info("Server %s started successfully", s.name)
	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("server %s is not running", s.name)
	}

	s.logger.Info("Stopping server %s", s.name)
	s.running = false
	s.cancel()

	// Wait for goroutine to finish
	s.wg.Wait()

	s.logger.Info("Server %s stopped successfully", s.name)
	return nil
}

func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Server) GetName() string {
	return s.name
}

func (s *Server) GetPort() int {
	return s.port
}

func (s *Server) run() {
	defer s.wg.Done()

	s.logger.Debug("Server %s HTTP server goroutine started", s.name)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	// Start HTTP server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-s.ctx.Done()
	s.logger.Debug("Server %s received shutdown signal", s.name)

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error: %v", err)
	}
}

// WaitForShutdown blocks until the server is stopped
func (s *Server) WaitForShutdown() {
	s.wg.Wait()
}
