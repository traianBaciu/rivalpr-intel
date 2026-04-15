package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Journalist represents a media contact linked to an outlet.
// User-scoped: each user maintains their own journalist database.
// Supports soft delete.
// Composite unique: (user_id, email) — same journalist can exist for different users.
type Journalist struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_journalist_user_email" json:"user_id"`
	OutletID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"outlet_id"`
	Outlet    Outlet         `gorm:"foreignKey:OutletID;references:ID" json:"outlet,omitempty"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"not null;uniqueIndex:idx_journalist_user_email" json:"email"`
	Niche     string         `json:"niche"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (j *Journalist) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}
