package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ModelUsingUUID is like gorm.Model but with ID field changed to uuid.UUID
type ModelUsingUUID struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (model *ModelUsingUUID) BeforeCreate(_ *gorm.DB) (err error) {
	model.ID = uuid.New()
	return
}

// ModelWithoutID is like gorm.Model but without ID field.
// You must specify your own primaryKey
type ModelWithoutID struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
