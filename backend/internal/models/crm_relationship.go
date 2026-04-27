package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CrmRelationship stores a user's personal score and notes for a journalist.
// Composite unique index ensures one relationship record per (user, journalist).
// No soft delete — records are hard-deleted.
type CrmRelationship struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_crm_user_journalist" json:"user_id"`
	JournalistID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_crm_user_journalist" json:"journalist_id"`
	Journalist        Journalist `gorm:"foreignKey:JournalistID;references:ID" json:"journalist,omitempty"`
	RelationshipScore int        `gorm:"not null;default:0;check:chk_crm_score,relationship_score >= 0 AND relationship_score <= 10" json:"relationship_score"`
	PrivateNotes      string     `json:"private_notes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (r *CrmRelationship) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
