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

type updateOutletRequest struct {
	Name    string `json:"name"`
	Website string `json:"website"`
	Country string `json:"country"`
}

// ListOutlets handles GET /api/outlets
func (h *Handler) ListOutlets(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var outlets []models.Outlet
	if err := h.DB.Where("user_id = ?", userID).Find(&outlets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch outlets"})
		return
	}

	c.JSON(http.StatusOK, outlets)
}

// GetOutlet handles GET /api/outlets/:id
func (h *Handler) GetOutlet(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet ID"})
		return
	}

	var outlet models.Outlet
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&outlet).Error; err != nil {
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
func (h *Handler) CreateOutlet(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req createOutletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	outlet := models.Outlet{
		UserID:  userID,
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
func (h *Handler) UpdateOutlet(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet ID"})
		return
	}

	var outlet models.Outlet
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&outlet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "outlet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch outlet"})
		return
	}

	var req updateOutletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		outlet.Name = req.Name
	}
	if req.Website != "" {
		outlet.Website = req.Website
	}
	if req.Country != "" {
		outlet.Country = req.Country
	}

	if err := h.DB.Save(&outlet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update outlet"})
		return
	}

	c.JSON(http.StatusOK, outlet)
}

// DeleteOutlet handles DELETE /api/outlets/:id (soft delete)
func (h *Handler) DeleteOutlet(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet ID"})
		return
	}

	var outlet models.Outlet
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&outlet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "outlet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch outlet"})
		return
	}

	if err := h.DB.Delete(&outlet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete outlet"})
		return
	}

	c.Status(http.StatusNoContent)
}
