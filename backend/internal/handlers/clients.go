package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"gorm.io/gorm"
)

type createClientRequest struct {
	Name     string `json:"name" binding:"required"`
	Industry string `json:"industry"`
}

type updateClientRequest struct {
	Name     string `json:"name"`
	Industry string `json:"industry"`
}

// ListClients handles GET /api/clients
func (h *Handler) ListClients(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var clients []models.Client
	if err := h.DB.Where("user_id = ?", userID).Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch clients"})
		return
	}

	c.JSON(http.StatusOK, clients)
}

// GetClient handles GET /api/clients/:id
func (h *Handler) GetClient(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	var client models.Client
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch client"})
		return
	}

	c.JSON(http.StatusOK, client)
}

// CreateClient handles POST /api/clients
func (h *Handler) CreateClient(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req createClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := models.Client{
		UserID:   userID,
		Name:     req.Name,
		Industry: req.Industry,
	}

	if err := h.DB.Create(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create client"})
		return
	}

	c.JSON(http.StatusCreated, client)
}

// UpdateClient handles PUT /api/clients/:id
func (h *Handler) UpdateClient(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	var client models.Client
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch client"})
		return
	}

	var req updateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		client.Name = req.Name
	}
	if req.Industry != "" {
		client.Industry = req.Industry
	}

	if err := h.DB.Save(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update client"})
		return
	}

	c.JSON(http.StatusOK, client)
}

// DeleteClient handles DELETE /api/clients/:id (soft delete)
func (h *Handler) DeleteClient(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	var client models.Client
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch client"})
		return
	}

	if err := h.DB.Delete(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete client"})
		return
	}

	c.Status(http.StatusNoContent)
}
