package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Campaign represents a major press launch for a client.
// User-scoped. Supports soft delete.
type Campaign struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	ClientID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"client_id"`
	Client           Client         `gorm:"foreignKey:ClientID;references:ID" json:"client,omitempty"`
	Title            string         `gorm:"not null" json:"title"`
	PressReleaseText string         `json:"press_release_text"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Campaign) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
