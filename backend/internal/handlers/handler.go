package handlers

import (
	"github.com/rivalpr/backend/internal/config"
	"github.com/rivalpr/backend/internal/services"
	"gorm.io/gorm"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	DB          *gorm.DB
	Config      *config.Config
	AuthService *services.AuthService
}

// NewHandler constructs a Handler with the given dependencies.
func NewHandler(db *gorm.DB, cfg *config.Config, authService *services.AuthService) *Handler {
	return &Handler{
		DB:          db,
		Config:      cfg,
		AuthService: authService,
	}
}
