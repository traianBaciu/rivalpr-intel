package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Pitch represents an outreach message targeting a specific journalist.
// status is constrained to: draft | sent | opened | replied
// No soft delete — pitches are hard-deleted.
type Pitch struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ClientID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"client_id"`
	Client       Client     `gorm:"foreignKey:ClientID;references:ID" json:"client,omitempty"`
	CampaignID   *uuid.UUID `gorm:"type:uuid;index" json:"campaign_id"`
	Campaign     *Campaign  `gorm:"foreignKey:CampaignID;references:ID" json:"campaign,omitempty"`
	JournalistID uuid.UUID  `gorm:"type:uuid;not null;index" json:"journalist_id"`
	Journalist   Journalist `gorm:"foreignKey:JournalistID;references:ID" json:"journalist,omitempty"`
	ContextBrief string     `json:"context_brief"`
	Status       string     `gorm:"not null;default:'draft';check:chk_pitch_status,status IN ('draft','sent','opened','replied')" json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (p *Pitch) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
