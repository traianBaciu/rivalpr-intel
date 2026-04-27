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

// ListJournalists handles GET /api/journalists
// Returns only journalists the current user is tracking (has a CRM relationship with).
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

	query := h.DB.Preload("Outlet").
		Joins("JOIN crm_relationships cr ON cr.journalist_id = journalists.id AND cr.user_id = ?", userID)

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

// ListAgencyJournalists handles GET /api/journalists/agency
// Returns all agency journalists not yet tracked by the current user.
// Query params: ?page=1&limit=50&search=<str>&niche=<str>
func (h *Handler) ListAgencyJournalists(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	query := h.DB.Preload("Outlet").
		Where("id NOT IN (SELECT journalist_id FROM crm_relationships WHERE user_id = ?)", userID)

	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"
		query = query.Where("journalists.name ILIKE ? OR journalists.email ILIKE ?", like, like)
	}
	if niche := c.Query("niche"); niche != "" {
		query = query.Where("journalists.niche = ?", niche)
	}

	var total int64
	query.Model(&models.Journalist{}).Count(&total)

	var journalists []models.Journalist
	if err := query.Offset(offset).Limit(limit).Find(&journalists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch agency journalists"})
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journalist ID"})
		return
	}

	var journalist models.Journalist
	if err := h.DB.Preload("Outlet").First(&journalist, "id = ?", id).Error; err != nil {
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
// Adds a journalist to the agency-shared database. AddedBy is set from JWT.
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

	// Validate outlet exists in the agency database (no user ownership check).
	var outlet models.Outlet
	if err := h.DB.First(&outlet, "id = ?", outletID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outlet not found"})
		return
	}

	journalist := models.Journalist{
		AddedBy:  &userID,
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
// Journalists are agency-wide shared records and cannot be modified after creation.
func (h *Handler) UpdateJournalist(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "journalists are shared agency records and cannot be modified"})
}

// DeleteJournalist handles DELETE /api/journalists/:id
// Journalists are agency-wide shared records and cannot be deleted.
// To stop tracking a journalist, delete the CRM relationship instead.
func (h *Handler) DeleteJournalist(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "journalists are shared agency records and cannot be deleted; delete the CRM relationship to stop tracking"})
}
