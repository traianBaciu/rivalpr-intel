package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Client represents a brand represented by the agency.
// User-scoped: each user maintains their own client list.
// Supports soft delete.
type Client struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Name      string         `gorm:"not null" json:"name"`
	Industry  string         `json:"industry"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Client) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
