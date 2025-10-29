package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ontair/admin-panel/internal/adapters/primary/api"
	"github.com/ontair/admin-panel/internal/adapters/primary/middleware"
	"github.com/ontair/admin-panel/internal/adapters/secondary/cookie"
	userRepo "github.com/ontair/admin-panel/internal/adapters/secondary/database"
	"github.com/ontair/admin-panel/internal/adapters/secondary/jwt"
	"github.com/ontair/admin-panel/internal/core/ports/service"
	"github.com/ontair/admin-panel/internal/core/services"
	"github.com/ontair/admin-panel/internal/infra/config"
	"github.com/ontair/admin-panel/internal/infra/database"
	"github.com/ontair/admin-panel/internal/infra/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Close()

	appLogger.Info("Starting Admin Panel Server")

	// Initialize database
	dbService, err := database.NewDatabaseService(cfg)
	if err != nil {
		appLogger.Fatal("Failed to initialize database")
	}
	defer dbService.Close()

	appLogger.Info("Database connection established")

	// Initialize dependencies
	deps := initializeDependencies(cfg, dbService, appLogger)

	// Setup Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := setupRouter(deps, cfg, appLogger)

	// Create server
	server := &http.Server{
		Addr:           cfg.GetPort(),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in background
	go func() {
		appLogger.Info("Server starting on port " + cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown")
	}

	appLogger.Info("Server exited")
}

// initializeDependencies sets up all application dependencies
func initializeDependencies(cfg *config.Config, dbService *database.DatabaseService, appLogger service.Logger) *Dependencies {
	// Initialize repositories
	userRepository := userRepo.NewUserRepository(dbService.GetPool())

	// Initialize external services
	jwtService := jwt.NewJWTService(cfg)
	cookieService := cookie.NewCookieService(cfg.Cookie.SameSite, cfg.Cookie.Domain, cfg.Cookie.Secure)

	// Initialize use cases
	authService := services.NewAuthService(userRepository, jwtService)
	userService := services.NewUserService(userRepository)

	return &Dependencies{
		Config:        cfg,
		Logger:        appLogger,
		AuthService:   authService,
		UserService:   userService,
		JWTService:    jwtService,
		CookieService: cookieService,
	}
}

// setupRouter configures Gin router with all routes and middleware
func setupRouter(deps *Dependencies, cfg *config.Config, appLogger service.Logger) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware - handled by Nginx proxy
	// No CORS headers needed here as Nginx handles them

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"time":    time.Now(),
			"version": "1.0.0",
		})
	})

	// API routes
	apiGroup := router.Group("/api/v1")

	// Initialize handlers
	authHandler := api.NewAuthHandler(deps.AuthService, appLogger, deps.CookieService, deps.JWTService)
	userHandler := api.NewUserHandler(deps.UserService, appLogger)

	// Init auth middleware
	authMiddleware := middleware.NewAuthMiddleware(deps.JWTService, appLogger, deps.CookieService, deps.AuthService)

	// Register auth routes (login, refresh, logout are public)
	authHandler.RegisterPublicRoutes(apiGroup)

	// Protected routes (require authentication)
	protected := apiGroup.Group("/")
	protected.Use(authMiddleware.RequireAuth())
	authHandler.RegisterProtectedRoutes(protected)
	userHandler.RegisterRoutes(protected)

	// Manager routes (require manager or higher role)
	manager := protected.Group("/manager")
	manager.Use(authMiddleware.RequireManagerOrHigher())
	authHandler.RegisterManagerRoutes(manager) // Register endpoint for manager+
	userHandler.RegisterManagerRoutes(manager) // User management for manager+

	// Admin routes (require admin role)
	admin := protected.Group("/admin")
	admin.Use(authMiddleware.RequireAdmin())
	userHandler.RegisterAdminRoutes(admin) // Admin-specific endpoints (full user list)

	return router
}

// Dependencies holds all application dependencies
type Dependencies struct {
	Config        *config.Config
	Logger        service.Logger
	AuthService   service.AuthService
	UserService   service.UserService
	JWTService    service.JWTService
	CookieService service.CookieService
}
