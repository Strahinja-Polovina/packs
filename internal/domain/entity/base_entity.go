package entity

import (
	"github.com/google/uuid"
	"time"
)

// BaseEntity for easier manage entities
type BaseEntity struct {
	id        uuid.UUID
	createdAt time.Time
	updatedAt time.Time
}

func NewBaseEntity(id uuid.UUID) BaseEntity {
	now := time.Now()
	return BaseEntity{
		id:        id,
		createdAt: now,
		updatedAt: now,
	}
}

func (b *BaseEntity) ID() uuid.UUID {
	return b.id
}

func (b *BaseEntity) CreatedAt() time.Time {
	return b.createdAt
}

func (b *BaseEntity) UpdatedAt() time.Time {
	return b.updatedAt
}

func (b *BaseEntity) Update() {
	b.updatedAt = time.Now()
}

// SetTimestamps sets the created and updated timestamps from database values
func (b *BaseEntity) SetTimestamps(createdAt, updatedAt time.Time) {
	b.createdAt = createdAt
	b.updatedAt = updatedAt
}
