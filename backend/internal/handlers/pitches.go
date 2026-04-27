package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"gorm.io/gorm"
)

var validStatuses = map[string]bool{
	"draft":   true,
	"sent":    true,
	"opened":  true,
	"replied": true,
}

type createPitchRequest struct {
	ClientID     string  `json:"client_id" binding:"required"`
	JournalistID string  `json:"journalist_id" binding:"required"`
	CampaignID   *string `json:"campaign_id"`
	ContextBrief string  `json:"context_brief"`
}

type updatePitchRequest struct {
	Status       string `json:"status"`
	ContextBrief string `json:"context_brief"`
}

// ListPitches handles GET /api/pitches
// Query params: ?campaign_id=<uuid>&status=<draft|sent|opened|replied>
func (h *Handler) ListPitches(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	query := h.DB.Preload("Client").Preload("Journalist").Preload("Journalist.Outlet").
		Preload("Campaign").Where("pitches.user_id = ?", userID)

	if campaignIDStr := c.Query("campaign_id"); campaignIDStr != "" {
		campaignID, err := uuid.Parse(campaignIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign_id"})
			return
		}
		query = query.Where("pitches.campaign_id = ?", campaignID)
	}

	if status := c.Query("status"); status != "" {
		if !validStatuses[status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status; must be one of: draft, sent, opened, replied"})
			return
		}
		query = query.Where("pitches.status = ?", status)
	}

	var pitches []models.Pitch
	if err := query.Find(&pitches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pitches"})
		return
	}

	c.JSON(http.StatusOK, pitches)
}

// GetPitch handles GET /api/pitches/:id
func (h *Handler) GetPitch(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pitch ID"})
		return
	}

	var pitch models.Pitch
	if err := h.DB.
		Preload("Client").
		Preload("Journalist").
		Preload("Journalist.Outlet").
		Preload("Campaign").
		Where("id = ? AND user_id = ?", id, userID).First(&pitch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "pitch not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pitch"})
		return
	}

	c.JSON(http.StatusOK, pitch)
}

// CreatePitch handles POST /api/pitches
func (h *Handler) CreatePitch(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req createPitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}
	journalistID, err := uuid.Parse(req.JournalistID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journalist_id"})
		return
	}

	// Validate ownership of client and journalist.
	var client models.Client
	if err := h.DB.Where("id = ? AND user_id = ?", clientID, userID).First(&client).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client not found"})
		return
	}
	var journalist models.Journalist
	if err := h.DB.First(&journalist, "id = ?", journalistID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "journalist not found"})
		return
	}

	pitch := models.Pitch{
		UserID:       userID,
		ClientID:     clientID,
		JournalistID: journalistID,
		ContextBrief: req.ContextBrief,
		Status:       "draft",
	}

	if req.CampaignID != nil {
		campaignID, err := uuid.Parse(*req.CampaignID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign_id"})
			return
		}
		// Validate campaign belongs to requesting user.
		var campaign models.Campaign
		if err := h.DB.Where("id = ? AND user_id = ?", campaignID, userID).First(&campaign).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "campaign not found"})
			return
		}
		pitch.CampaignID = &campaignID
	}

	if err := h.DB.Create(&pitch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create pitch"})
		return
	}

	h.DB.Preload("Client").Preload("Journalist").Preload("Journalist.Outlet").Preload("Campaign").First(&pitch, pitch.ID)
	c.JSON(http.StatusCreated, pitch)
}

// UpdatePitch handles PUT /api/pitches/:id
func (h *Handler) UpdatePitch(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pitch ID"})
		return
	}

	var pitch models.Pitch
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&pitch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "pitch not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pitch"})
		return
	}

	var req updatePitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != "" {
		if !validStatuses[req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status; must be one of: draft, sent, opened, replied"})
			return
		}
		pitch.Status = req.Status
	}
	if req.ContextBrief != "" {
		pitch.ContextBrief = req.ContextBrief
	}

	if err := h.DB.Save(&pitch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update pitch"})
		return
	}

	h.DB.Preload("Client").Preload("Journalist").Preload("Journalist.Outlet").Preload("Campaign").First(&pitch, pitch.ID)
	c.JSON(http.StatusOK, pitch)
}

// DeletePitch handles DELETE /api/pitches/:id (hard delete)
func (h *Handler) DeletePitch(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pitch ID"})
		return
	}

	result := h.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Pitch{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete pitch"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "pitch not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
