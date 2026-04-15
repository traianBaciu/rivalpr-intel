package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Outlet represents a media publication (e.g. TechCrunch, Forbes).
// User-scoped: each user maintains their own outlet list.
// Supports soft delete.
type Outlet struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_outlet_user_name" json:"user_id"`
	Name      string         `gorm:"not null;uniqueIndex:idx_outlet_user_name" json:"name"`
	Website   string         `json:"website"`
	Country   string         `json:"country"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (o *Outlet) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
