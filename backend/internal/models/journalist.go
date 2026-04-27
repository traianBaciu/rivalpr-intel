package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Journalist represents a media contact linked to an outlet.
// Agency-wide shared record — visible to all authenticated users.
// AddedBy tracks who created the record. Email is globally unique.
// Supports soft delete.
type Journalist struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	AddedBy   *uuid.UUID     `gorm:"type:uuid;index" json:"added_by"`
	OutletID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"outlet_id"`
	Outlet    Outlet         `gorm:"foreignKey:OutletID;references:ID" json:"outlet,omitempty"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"not null;uniqueIndex" json:"email"`
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
