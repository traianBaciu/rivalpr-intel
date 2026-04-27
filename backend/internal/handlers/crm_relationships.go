package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"gorm.io/gorm"
)

type createCrmRelationshipRequest struct {
	JournalistID      string `json:"journalist_id" binding:"required"`
	RelationshipScore int    `json:"relationship_score" binding:"omitempty,min=0,max=10"`
	PrivateNotes      string `json:"private_notes"`
}

type updateCrmRelationshipRequest struct {
	RelationshipScore int    `json:"relationship_score" binding:"omitempty,min=0,max=10"`
	PrivateNotes      string `json:"private_notes"`
}

// ListCrmRelationships handles GET /api/crm/relationships
func (h *Handler) ListCrmRelationships(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var relationships []models.CrmRelationship
	if err := h.DB.Preload("Journalist").Preload("Journalist.Outlet").
		Where("user_id = ?", userID).Find(&relationships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch relationships"})
		return
	}

	c.JSON(http.StatusOK, relationships)
}

// GetCrmRelationship handles GET /api/crm/relationships/:id
func (h *Handler) GetCrmRelationship(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid relationship ID"})
		return
	}

	var rel models.CrmRelationship
	if err := h.DB.Preload("Journalist").Preload("Journalist.Outlet").
		Where("id = ? AND user_id = ?", id, userID).First(&rel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "relationship not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch relationship"})
		return
	}

	c.JSON(http.StatusOK, rel)
}

// CreateCrmRelationship handles POST /api/crm/relationships
func (h *Handler) CreateCrmRelationship(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req createCrmRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journalistID, err := uuid.Parse(req.JournalistID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journalist_id"})
		return
	}

	// Journalist is an agency-wide shared record — no user ownership check needed.
	var journalist models.Journalist
	if err := h.DB.First(&journalist, "id = ?", journalistID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "journalist not found"})
		return
	}

	rel := models.CrmRelationship{
		UserID:            userID,
		JournalistID:      journalistID,
		RelationshipScore: req.RelationshipScore,
		PrivateNotes:      req.PrivateNotes,
	}

	if err := h.DB.Create(&rel).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "relationship already exists for this journalist; use PUT to update"})
		return
	}

	rel.Journalist = journalist
	c.JSON(http.StatusCreated, rel)
}

// UpdateCrmRelationship handles PUT /api/crm/relationships/:id
func (h *Handler) UpdateCrmRelationship(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid relationship ID"})
		return
	}

	var rel models.CrmRelationship
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&rel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "relationship not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch relationship"})
		return
	}

	var req updateCrmRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RelationshipScore != 0 {
		rel.RelationshipScore = req.RelationshipScore
	}
	rel.PrivateNotes = req.PrivateNotes

	if err := h.DB.Save(&rel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update relationship"})
		return
	}

	h.DB.Preload("Journalist").Preload("Journalist.Outlet").First(&rel, rel.ID)
	c.JSON(http.StatusOK, rel)
}

// DeleteCrmRelationship handles DELETE /api/crm/relationships/:id (hard delete)
func (h *Handler) DeleteCrmRelationship(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid relationship ID"})
		return
	}

	result := h.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.CrmRelationship{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete relationship"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "relationship not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
