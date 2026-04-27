package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PitchVersion stores each AI generation attempt for a pitch.
// Composite unique: (pitch_id, version_number) — version numbers are sequential per pitch.
// No soft delete.
type PitchVersion struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PitchID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_pitch_version" json:"pitch_id"`
	VersionNumber    int       `gorm:"not null;uniqueIndex:uq_pitch_version" json:"version_number"`
	AIGeneratedBody  string    `gorm:"not null" json:"ai_generated_body"`
	PromptSnapshot   string    `json:"prompt_snapshot"`
	GenerationParams string    `json:"generation_params"` // JSON: tone, length, angle, custom_instructions, reference info
	CreatedAt        time.Time `json:"created_at"`
}

func (pv *PitchVersion) BeforeCreate(tx *gorm.DB) error {
	if pv.ID == uuid.Nil {
		pv.ID = uuid.New()
	}
	return nil
}
