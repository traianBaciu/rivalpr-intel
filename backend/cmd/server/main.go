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
	"github.com/joho/godotenv"
	"github.com/rivalpr/backend/internal/config"
	"github.com/rivalpr/backend/internal/handlers"
	"github.com/rivalpr/backend/internal/middleware"
	"github.com/rivalpr/backend/internal/models"
	"github.com/rivalpr/backend/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file in development (ignored if not present in production).
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	// Run auto-migrations for all 8 models.
	if err := db.AutoMigrate(
		&models.User{},
		&models.Outlet{},
		&models.Client{},
		&models.Journalist{},
		&models.CrmRelationship{},
		&models.Campaign{},
		&models.Pitch{},
		&models.PitchVersion{},
	); err != nil {
		log.Fatalf("auto-migration failed: %v", err)
	}
	log.Println("Database migrations applied")

	// Initialise services and handler.
	authService := services.NewAuthService(db, cfg)
	h := handlers.NewHandler(db, cfg, authService)

	// Rate limiters: auth (10/min by IP), AI (from config, by user), general (120/min by user).
	authLimiter := middleware.NewRateLimiter(10)
	aiLimiter := middleware.NewRateLimiter(cfg.AIRateLimitRPM)
	generalLimiter := middleware.NewRateLimiter(120)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check (no auth, no rate limit).
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().UTC()})
	})

	// ── Public auth routes (IP-based rate limit: 10 req/min) ──────────────────
	auth := router.Group("/auth")
	auth.Use(authLimiter.Limit(middleware.IPKeyFunc))
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}

	// ── Protected API routes (JWT required, 120 req/min per user) ─────────────
	api := router.Group("/api")
	api.Use(middleware.AuthRequired(authService))
	api.Use(generalLimiter.Limit(middleware.UserKeyFunc))
	{
		// Clients
		clients := api.Group("/clients")
		{
			clients.GET("", h.ListClients)
			clients.GET("/:id", h.GetClient)
			clients.POST("", h.CreateClient)
			clients.PUT("/:id", h.UpdateClient)
			clients.DELETE("/:id", h.DeleteClient)
		}

		// Outlets
		outlets := api.Group("/outlets")
		{
			outlets.GET("", h.ListOutlets)
			outlets.GET("/:id", h.GetOutlet)
			outlets.POST("", h.CreateOutlet)
			outlets.PUT("/:id", h.UpdateOutlet)
			outlets.DELETE("/:id", h.DeleteOutlet)
		}

		// Journalists (paginated)
		journalists := api.Group("/journalists")
		{
			journalists.GET("", h.ListJournalists)
			journalists.GET("/:id", h.GetJournalist)
			journalists.POST("", h.CreateJournalist)
			journalists.PUT("/:id", h.UpdateJournalist)
			journalists.DELETE("/:id", h.DeleteJournalist)
		}

		// CRM Relationships
		crm := api.Group("/crm/relationships")
		{
			crm.GET("", h.ListCrmRelationships)
			crm.GET("/:id", h.GetCrmRelationship)
			crm.POST("", h.CreateCrmRelationship)
			crm.PUT("/:id", h.UpdateCrmRelationship)
			crm.DELETE("/:id", h.DeleteCrmRelationship)
		}

		// Campaigns
		campaigns := api.Group("/campaigns")
		{
			campaigns.GET("", h.ListCampaigns)
			campaigns.GET("/:id", h.GetCampaign)
			campaigns.POST("", h.CreateCampaign)
			campaigns.PUT("/:id", h.UpdateCampaign)
			campaigns.DELETE("/:id", h.DeleteCampaign)
		}

		// Pitches + versions + AI generation
		pitches := api.Group("/pitches")
		{
			pitches.GET("", h.ListPitches)
			pitches.GET("/:id", h.GetPitch)
			pitches.POST("", h.CreatePitch)
			pitches.PUT("/:id", h.UpdatePitch)
			pitches.DELETE("/:id", h.DeletePitch)
			pitches.GET("/:id/versions", h.ListPitchVersions)
			pitches.GET("/:id/versions/:versionId", h.GetPitchVersion)
			// AI endpoint has an additional per-user AI rate limit on top of the general limiter.
			pitches.POST("/:id/generate", aiLimiter.Limit(middleware.UserKeyFunc), h.GeneratePitch)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("RivalPR Intel backend listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("Server exited cleanly")
}

