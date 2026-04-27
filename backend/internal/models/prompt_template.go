package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PromptTemplate stores reusable AI generation presets for a user.
// User-scoped. Supports soft delete.
type PromptTemplate struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:uq_template_user_name" json:"user_id"`
	Name               string         `gorm:"not null;uniqueIndex:uq_template_user_name" json:"name"`
	Tone               string         `json:"tone"`
	Length             string         `json:"length"`
	Angle              string         `json:"angle"`
	CustomInstructions string         `json:"custom_instructions"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (pt *PromptTemplate) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	return nil
}
