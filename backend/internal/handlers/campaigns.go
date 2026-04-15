package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"gorm.io/gorm"
)

type createCampaignRequest struct {
	ClientID         string `json:"client_id" binding:"required"`
	Title            string `json:"title" binding:"required"`
	PressReleaseText string `json:"press_release_text"`
}

type updateCampaignRequest struct {
	Title            string `json:"title"`
	PressReleaseText string `json:"press_release_text"`
}

// ListCampaigns handles GET /api/campaigns
func (h *Handler) ListCampaigns(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var campaigns []models.Campaign
	if err := h.DB.Preload("Client").Where("user_id = ?", userID).Find(&campaigns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch campaigns"})
		return
	}

	c.JSON(http.StatusOK, campaigns)
}

// GetCampaign handles GET /api/campaigns/:id
func (h *Handler) GetCampaign(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign ID"})
		return
	}

	var campaign models.Campaign
	if err := h.DB.Preload("Client").Where("id = ? AND user_id = ?", id, userID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch campaign"})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

// CreateCampaign handles POST /api/campaigns
func (h *Handler) CreateCampaign(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	// Validate client belongs to requesting user.
	var client models.Client
	if err := h.DB.Where("id = ? AND user_id = ?", clientID, userID).First(&client).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client not found"})
		return
	}

	campaign := models.Campaign{
		UserID:           userID,
		ClientID:         clientID,
		Title:            req.Title,
		PressReleaseText: req.PressReleaseText,
	}

	if err := h.DB.Create(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create campaign"})
		return
	}

	campaign.Client = client
	c.JSON(http.StatusCreated, campaign)
}

// UpdateCampaign handles PUT /api/campaigns/:id
func (h *Handler) UpdateCampaign(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign ID"})
		return
	}

	var campaign models.Campaign
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch campaign"})
		return
	}

	var req updateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		campaign.Title = req.Title
	}
	campaign.PressReleaseText = req.PressReleaseText

	if err := h.DB.Save(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update campaign"})
		return
	}

	h.DB.Preload("Client").First(&campaign, campaign.ID)
	c.JSON(http.StatusOK, campaign)
}

// DeleteCampaign handles DELETE /api/campaigns/:id (soft delete)
func (h *Handler) DeleteCampaign(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign ID"})
		return
	}

	var campaign models.Campaign
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch campaign"})
		return
	}

	if err := h.DB.Delete(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete campaign"})
		return
	}

	c.Status(http.StatusNoContent)
}
