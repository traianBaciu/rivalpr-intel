package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"github.com/rivalpr/backend/internal/worker"
	"gorm.io/gorm"
)

// ListPitchVersions handles GET /api/pitches/:id/versions
func (h *Handler) ListPitchVersions(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	pitchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pitch ID"})
		return
	}

	// Verify pitch belongs to requesting user.
	var pitch models.Pitch
	if err := h.DB.Where("id = ? AND user_id = ?", pitchID, userID).First(&pitch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "pitch not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pitch"})
		return
	}

	var versions []models.PitchVersion
	if err := h.DB.Where("pitch_id = ?", pitchID).Order("version_number ASC").Find(&versions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pitch versions"})
		return
	}

	c.JSON(http.StatusOK, versions)
}

// GetPitchVersion handles GET /api/pitches/:id/versions/:versionId
func (h *Handler) GetPitchVersion(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	pitchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pitch ID"})
		return
	}

	versionID, err := uuid.Parse(c.Param("versionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version ID"})
		return
	}

	// Verify pitch belongs to requesting user.
	var pitch models.Pitch
	if err := h.DB.Where("id = ? AND user_id = ?", pitchID, userID).First(&pitch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "pitch not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pitch"})
		return
	}

	var version models.PitchVersion
	if err := h.DB.Where("id = ? AND pitch_id = ?", versionID, pitchID).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch version"})
		return
	}

	c.JSON(http.StatusOK, version)
}

// GeneratePitch handles POST /api/pitches/:id/generate
// Enqueues an async AI generation job and returns 202 Accepted.
// The client should poll GET /api/pitches/:id/versions to retrieve the result.
func (h *Handler) GeneratePitch(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	pitchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pitch ID"})
		return
	}

	// Load pitch with all associations needed for context injection.
	var pitch models.Pitch
	if err := h.DB.
		Preload("Client").
		Preload("Journalist").
		Preload("Journalist.Outlet").
		Preload("Campaign").
		Where("id = ? AND user_id = ?", pitchID, userID).First(&pitch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "pitch not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pitch"})
		return
	}

	// Determine the next version number.
	var maxVersion int
	h.DB.Model(&models.PitchVersion{}).
		Where("pitch_id = ?", pitchID).
		Select("COALESCE(MAX(version_number), 0)").
		Scan(&maxVersion)
	nextVersion := maxVersion + 1

	// Build context strings.
	outletName := pitch.Journalist.Outlet.Name
	if outletName == "" {
		outletName = "their publication"
	}
	niche := pitch.Journalist.Niche
	if niche == "" {
		niche = "general"
	}

	campaignTitle := ""
	contextText := pitch.ContextBrief
	if pitch.Campaign != nil {
		campaignTitle = pitch.Campaign.Title
		if contextText == "" {
			text := pitch.Campaign.PressReleaseText
			if len(text) > 300 {
				text = text[:300] + "..."
			}
			contextText = text
		}
	}
	if contextText == "" {
		contextText = fmt.Sprintf(
			"%s is launching an exciting new initiative that would interest %s readers.",
			pitch.Client.Name, niche,
		)
	}

	// Look up CRM relationship score for tone calibration.
	var crmScore int
	var rel models.CrmRelationship
	if err := h.DB.Where("user_id = ? AND journalist_id = ?", userID, pitch.JournalistID).
		First(&rel).Error; err == nil {
		crmScore = rel.RelationshipScore
	}

	// Compile the prompt snapshot stored alongside the version for auditability.
	snapshot := fmt.Sprintf(
		"Journalist: %s @ %s (niche: %s) | Client: %s | Campaign: %s | CRM score: %d/10 | Context: %s",
		pitch.Journalist.Name, outletName, niche, pitch.Client.Name, campaignTitle, crmScore, contextText,
	)

	job := worker.PitchJob{
		PitchID:        pitchID,
		VersionNumber:  nextVersion,
		PromptSnapshot: snapshot,
		JournalistName: pitch.Journalist.Name,
		OutletName:     outletName,
		Niche:          niche,
		ClientName:     pitch.Client.Name,
		CampaignTitle:  campaignTitle,
		ContextText:    contextText,
		CRMScore:       crmScore,
	}

	if !h.Worker.Enqueue(job) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "generation queue is full, please retry shortly"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":        "generation queued",
		"pitch_id":       pitchID,
		"version_number": nextVersion,
	})
}
