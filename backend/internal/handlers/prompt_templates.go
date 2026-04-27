package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"gorm.io/gorm"
)

var validTones = map[string]bool{"formal": true, "conversational": true, "urgent": true, "enthusiastic": true}
var validLengths = map[string]bool{"concise": true, "standard": true, "detailed": true}
var validAngles = map[string]bool{"news_hook": true, "exclusive": true, "follow_up": true, "thought_leadership": true, "event": true}

type createPromptTemplateRequest struct {
	Name               string `json:"name" binding:"required"`
	Tone               string `json:"tone"`
	Length             string `json:"length"`
	Angle              string `json:"angle"`
	CustomInstructions string `json:"custom_instructions"`
}

type updatePromptTemplateRequest struct {
	Name               string `json:"name"`
	Tone               string `json:"tone"`
	Length             string `json:"length"`
	Angle              string `json:"angle"`
	CustomInstructions string `json:"custom_instructions"`
}

func validateTemplateFields(tone, length, angle string) string {
	if tone != "" && !validTones[tone] {
		return "invalid tone: must be one of formal, conversational, urgent, enthusiastic"
	}
	if length != "" && !validLengths[length] {
		return "invalid length: must be one of concise, standard, detailed"
	}
	if angle != "" && !validAngles[angle] {
		return "invalid angle: must be one of news_hook, exclusive, follow_up, thought_leadership, event"
	}
	return ""
}

// ListPromptTemplates handles GET /api/prompt-templates
func (h *Handler) ListPromptTemplates(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var templates []models.PromptTemplate
	if err := h.DB.Where("user_id = ?", userID).Order("name ASC").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch prompt templates"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// GetPromptTemplate handles GET /api/prompt-templates/:id
func (h *Handler) GetPromptTemplate(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var tmpl models.PromptTemplate
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&tmpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "prompt template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch prompt template"})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

// CreatePromptTemplate handles POST /api/prompt-templates
func (h *Handler) CreatePromptTemplate(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req createPromptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg := validateTemplateFields(req.Tone, req.Length, req.Angle); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	tmpl := models.PromptTemplate{
		UserID:             userID,
		Name:               req.Name,
		Tone:               req.Tone,
		Length:             req.Length,
		Angle:              req.Angle,
		CustomInstructions: req.CustomInstructions,
	}

	if err := h.DB.Create(&tmpl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create prompt template"})
		return
	}

	c.JSON(http.StatusCreated, tmpl)
}

// UpdatePromptTemplate handles PUT /api/prompt-templates/:id
func (h *Handler) UpdatePromptTemplate(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var tmpl models.PromptTemplate
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&tmpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "prompt template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch prompt template"})
		return
	}

	var req updatePromptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg := validateTemplateFields(req.Tone, req.Length, req.Angle); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if req.Name != "" {
		tmpl.Name = req.Name
	}
	tmpl.Tone = req.Tone
	tmpl.Length = req.Length
	tmpl.Angle = req.Angle
	tmpl.CustomInstructions = req.CustomInstructions

	if err := h.DB.Save(&tmpl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update prompt template"})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

// DeletePromptTemplate handles DELETE /api/prompt-templates/:id (soft delete)
func (h *Handler) DeletePromptTemplate(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var tmpl models.PromptTemplate
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&tmpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "prompt template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch prompt template"})
		return
	}

	if err := h.DB.Delete(&tmpl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete prompt template"})
		return
	}

	c.Status(http.StatusNoContent)
}
