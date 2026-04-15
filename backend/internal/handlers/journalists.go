package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"gorm.io/gorm"
)

type createJournalistRequest struct {
	OutletID string `json:"outlet_id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Niche    string `json:"niche"`
}

type updateJournalistRequest struct {
	OutletID string `json:"outlet_id"`
	Name     string `json:"name"`
	Email    string `json:"email" binding:"omitempty,email"`
	Niche    string `json:"niche"`
}

// ListJournalists handles GET /api/journalists
// Query params: ?page=1&limit=20&outlet_id=<uuid>
func (h *Handler) ListJournalists(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := h.DB.Where("journalists.user_id = ?", userID).Preload("Outlet")

	if outletIDStr := c.Query("outlet_id"); outletIDStr != "" {
		outletID, err := uuid.Parse(outletIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet_id"})
			return
		}
		query = query.Where("journalists.outlet_id = ?", outletID)
	}

	var total int64
	query.Model(&models.Journalist{}).Count(&total)

	var journalists []models.Journalist
	if err := query.Offset(offset).Limit(limit).Find(&journalists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch journalists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  journalists,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetJournalist handles GET /api/journalists/:id
func (h *Handler) GetJournalist(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journalist ID"})
		return
	}

	var journalist models.Journalist
	if err := h.DB.Preload("Outlet").Where("id = ? AND user_id = ?", id, userID).First(&journalist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "journalist not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch journalist"})
		return
	}

	c.JSON(http.StatusOK, journalist)
}

// CreateJournalist handles POST /api/journalists
func (h *Handler) CreateJournalist(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req createJournalistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	outletID, err := uuid.Parse(req.OutletID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet_id"})
		return
	}

	// Validate outlet belongs to requesting user.
	var outlet models.Outlet
	if err := h.DB.Where("id = ? AND user_id = ?", outletID, userID).First(&outlet).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outlet not found"})
		return
	}

	journalist := models.Journalist{
		UserID:   userID,
		OutletID: outletID,
		Name:     req.Name,
		Email:    req.Email,
		Niche:    req.Niche,
	}

	if err := h.DB.Create(&journalist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create journalist"})
		return
	}

	journalist.Outlet = outlet
	c.JSON(http.StatusCreated, journalist)
}

// UpdateJournalist handles PUT /api/journalists/:id
func (h *Handler) UpdateJournalist(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journalist ID"})
		return
	}

	var journalist models.Journalist
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&journalist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "journalist not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch journalist"})
		return
	}

	var req updateJournalistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OutletID != "" {
		outletID, err := uuid.Parse(req.OutletID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet_id"})
			return
		}
		// Validate new outlet belongs to requesting user.
		var outlet models.Outlet
		if err := h.DB.Where("id = ? AND user_id = ?", outletID, userID).First(&outlet).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "outlet not found"})
			return
		}
		journalist.OutletID = outletID
	}
	if req.Name != "" {
		journalist.Name = req.Name
	}
	if req.Email != "" {
		journalist.Email = req.Email
	}
	if req.Niche != "" {
		journalist.Niche = req.Niche
	}

	if err := h.DB.Save(&journalist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update journalist"})
		return
	}

	h.DB.Preload("Outlet").First(&journalist, journalist.ID)
	c.JSON(http.StatusOK, journalist)
}

// DeleteJournalist handles DELETE /api/journalists/:id (soft delete)
func (h *Handler) DeleteJournalist(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journalist ID"})
		return
	}

	var journalist models.Journalist
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&journalist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "journalist not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch journalist"})
		return
	}

	if err := h.DB.Delete(&journalist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete journalist"})
		return
	}

	c.Status(http.StatusNoContent)
}
