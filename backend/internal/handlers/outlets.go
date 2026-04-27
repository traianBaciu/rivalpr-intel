package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"gorm.io/gorm"
)

type createOutletRequest struct {
	Name    string `json:"name" binding:"required"`
	Website string `json:"website"`
	Country string `json:"country"`
}

// ListOutlets handles GET /api/outlets
// Returns all agency outlets (shared, not user-scoped).
func (h *Handler) ListOutlets(c *gin.Context) {
	var outlets []models.Outlet
	if err := h.DB.Find(&outlets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch outlets"})
		return
	}

	c.JSON(http.StatusOK, outlets)
}

// GetOutlet handles GET /api/outlets/:id
func (h *Handler) GetOutlet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet ID"})
		return
	}

	var outlet models.Outlet
	if err := h.DB.First(&outlet, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "outlet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch outlet"})
		return
	}

	c.JSON(http.StatusOK, outlet)
}

// CreateOutlet handles POST /api/outlets
// Adds a new outlet to the agency-shared database. AddedBy is set from JWT.
func (h *Handler) CreateOutlet(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req createOutletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	outlet := models.Outlet{
		AddedBy: &userID,
		Name:    req.Name,
		Website: req.Website,
		Country: req.Country,
	}

	if err := h.DB.Create(&outlet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create outlet"})
		return
	}

	c.JSON(http.StatusCreated, outlet)
}

// UpdateOutlet handles PUT /api/outlets/:id
// Outlets are agency-wide shared records and cannot be modified after creation.
func (h *Handler) UpdateOutlet(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "outlets are shared agency records and cannot be modified"})
}

// DeleteOutlet handles DELETE /api/outlets/:id
// Outlets are agency-wide shared records and cannot be deleted.
func (h *Handler) DeleteOutlet(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "outlets are shared agency records and cannot be deleted"})
}
