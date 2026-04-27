package handlers

import (
	"encoding/json"
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

// generateRequest holds optional AI generation controls.
// All fields are optional — an empty body produces the same result as the original endpoint.
type generateRequest struct {
	Tone               string `json:"tone"`
	Length             string `json:"length"`
	Angle              string `json:"angle"`
	CustomInstructions string `json:"custom_instructions"`
	ReferenceVersionID string `json:"reference_version_id"`
	RefinementNote     string `json:"refinement_note"`
}

// generationParamsJSON is serialized and stored on the PitchVersion for auditability.
type generationParamsJSON struct {
	Tone               string `json:"tone,omitempty"`
	Length             string `json:"length,omitempty"`
	Angle              string `json:"angle,omitempty"`
	CustomInstructions string `json:"custom_instructions,omitempty"`
	ReferenceVersionID string `json:"reference_version_id,omitempty"`
	RefinementNote     string `json:"refinement_note,omitempty"`
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

	// Bind optional generation controls (empty body is fine).
	var req generateRequest
	_ = c.ShouldBindJSON(&req)

	// Validate enum values.
	if req.Tone != "" && !validTones[req.Tone] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tone: must be one of formal, conversational, urgent, enthusiastic"})
		return
	}
	if req.Length != "" && !validLengths[req.Length] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid length: must be one of concise, standard, detailed"})
		return
	}
	if req.Angle != "" && !validAngles[req.Angle] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid angle: must be one of news_hook, exclusive, follow_up, thought_leadership, event"})
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
	if req.Tone != "" {
		snapshot += fmt.Sprintf(" | Tone: %s", req.Tone)
	}
	if req.Length != "" {
		snapshot += fmt.Sprintf(" | Length: %s", req.Length)
	}
	if req.Angle != "" {
		snapshot += fmt.Sprintf(" | Angle: %s", req.Angle)
	}
	if req.CustomInstructions != "" {
		snapshot += fmt.Sprintf(" | Instructions: %s", req.CustomInstructions)
	}

	// Resolve reference version body if refining a previous version.
	var referenceBody string
	if req.ReferenceVersionID != "" {
		refID, err := uuid.Parse(req.ReferenceVersionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reference_version_id"})
			return
		}
		var refVersion models.PitchVersion
		if err := h.DB.Where("id = ? AND pitch_id = ?", refID, pitchID).First(&refVersion).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "reference version not found for this pitch"})
			return
		}
		referenceBody = refVersion.AIGeneratedBody
		snapshot += fmt.Sprintf(" | Refining v%d", refVersion.VersionNumber)
	}

	// Serialize generation params as JSON for the version record.
	paramsJSON, _ := json.Marshal(generationParamsJSON{
		Tone:               req.Tone,
		Length:             req.Length,
		Angle:              req.Angle,
		CustomInstructions: req.CustomInstructions,
		ReferenceVersionID: req.ReferenceVersionID,
		RefinementNote:     req.RefinementNote,
	})

	// Load sender email for sign-off.
	var user models.User
	senderEmail := ""
	if err := h.DB.First(&user, "id = ?", userID).Error; err == nil {
		senderEmail = user.Email
	}

	job := worker.PitchJob{
		PitchID:            pitchID,
		VersionNumber:      nextVersion,
		PromptSnapshot:     snapshot,
		GenerationParams:   string(paramsJSON),
		JournalistName:     pitch.Journalist.Name,
		OutletName:         outletName,
		Niche:              niche,
		ClientName:         pitch.Client.Name,
		CampaignTitle:      campaignTitle,
		ContextText:        contextText,
		CRMScore:           crmScore,
		SenderEmail:        senderEmail,
		Tone:               req.Tone,
		Length:             req.Length,
		Angle:              req.Angle,
		CustomInstructions: req.CustomInstructions,
		RefinementNote:     req.RefinementNote,
		ReferenceBody:      referenceBody,
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
