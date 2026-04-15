package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
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
// Phase 1: returns a mock AI-generated pitch body so the full Epic 4 flow is testable.
// Phase 3: replace mockGenerate with real Anthropic Claude API call.
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
	journalistName := pitch.Journalist.Name
	outletName := pitch.Journalist.Outlet.Name
	if outletName == "" {
		outletName = "their publication"
	}
	niche := pitch.Journalist.Niche
	if niche == "" {
		niche = "general"
	}
	clientName := pitch.Client.Name

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
		contextText = fmt.Sprintf("%s is launching an exciting new initiative that would interest %s readers.", clientName, niche)
	}

	// Compile the prompt snapshot (what would be sent to the real AI).
	prompt := fmt.Sprintf(
		"Generate a personalized pitch email for journalist %s at %s (niche: %s) on behalf of client %s.",
		journalistName, outletName, niche, clientName,
	)
	if campaignTitle != "" {
		prompt += fmt.Sprintf(" Campaign context: %s.", campaignTitle)
	}
	prompt += fmt.Sprintf(" Key context: %s", contextText)

	// Mock AI response — template-generated to simulate real output.
	body := fmt.Sprintf(`Subject: Exclusive Story Opportunity for %s — %s

Dear %s,

I hope this message finds you well. I'm reaching out on behalf of %s with a story that I believe would resonate strongly with your %s readers at %s.

%s

This is a unique opportunity that aligns perfectly with the stories your audience cares about. I'd love to set up a 15-minute call to walk you through the details and answer any questions you might have.

Are you available for a brief conversation this week? I'm happy to work around your schedule.

Best regards,
[Your Name]
[Agency Name]
[Phone Number]

P.S. I've prepared an exclusive press kit with additional data and visuals — happy to share upon request.

---
[Mock AI Generation — v%d | Real Anthropic Claude integration coming in Phase 3]`,
		outletName, clientName,
		journalistName,
		clientName,
		niche, outletName,
		contextText,
		nextVersion,
	)

	version := models.PitchVersion{
		PitchID:         pitchID,
		VersionNumber:   nextVersion,
		AIGeneratedBody: body,
		PromptSnapshot:  prompt,
	}

	if err := h.DB.Create(&version).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save generated pitch version"})
		return
	}

	c.JSON(http.StatusCreated, version)
}
