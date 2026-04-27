package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Outlet represents a media publication (e.g. TechCrunch, Forbes).
// Agency-wide shared record — visible to all authenticated users.
// AddedBy tracks who created the record. Supports soft delete.
type Outlet struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	AddedBy   *uuid.UUID     `gorm:"type:uuid;index" json:"added_by"`
	Name      string         `gorm:"not null;uniqueIndex" json:"name"`
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
