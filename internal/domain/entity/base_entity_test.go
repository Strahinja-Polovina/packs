package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewBaseEntity(t *testing.T) {
	id := uuid.New()

	entity := NewBaseEntity(id)

	if entity.ID() != id {
		t.Errorf("Expected ID %s, got %s", id, entity.ID())
	}

	if entity.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be set, but it was zero")
	}

	if entity.UpdatedAt().IsZero() {
		t.Error("Expected UpdatedAt to be set, but it was zero")
	}

	if !entity.CreatedAt().Equal(entity.UpdatedAt()) {
		t.Error("Expected CreatedAt and UpdatedAt to be equal initially")
	}
}

func TestBaseEntity_ID(t *testing.T) {
	id := uuid.New()
	entity := NewBaseEntity(id)

	if entity.ID() != id {
		t.Errorf("Expected ID %s, got %s", id, entity.ID())
	}
}

func TestBaseEntity_CreatedAt(t *testing.T) {
	beforeCreation := time.Now()
	entity := NewBaseEntity(uuid.New())
	afterCreation := time.Now()

	createdAt := entity.CreatedAt()

	if createdAt.Before(beforeCreation) || createdAt.After(afterCreation) {
		t.Errorf("CreatedAt %v should be between %v and %v", createdAt, beforeCreation, afterCreation)
	}
}

func TestBaseEntity_UpdatedAt(t *testing.T) {
	entity := NewBaseEntity(uuid.New())
	initialUpdatedAt := entity.UpdatedAt()

	time.Sleep(1 * time.Millisecond)

	entity.Update()
	newUpdatedAt := entity.UpdatedAt()

	if !newUpdatedAt.After(initialUpdatedAt) {
		t.Errorf("UpdatedAt should be updated after calling Update(). Initial: %v, New: %v", initialUpdatedAt, newUpdatedAt)
	}
}

func TestBaseEntity_Update(t *testing.T) {
	entity := NewBaseEntity(uuid.New())
	originalCreatedAt := entity.CreatedAt()
	originalUpdatedAt := entity.UpdatedAt()

	time.Sleep(1 * time.Millisecond)

	entity.Update()

	if !entity.CreatedAt().Equal(originalCreatedAt) {
		t.Errorf("CreatedAt should not change after Update(). Original: %v, Current: %v", originalCreatedAt, entity.CreatedAt())
	}

	if !entity.UpdatedAt().After(originalUpdatedAt) {
		t.Errorf("UpdatedAt should be updated after calling Update(). Original: %v, Current: %v", originalUpdatedAt, entity.UpdatedAt())
	}
}

func TestBaseEntity_MultipleUpdates(t *testing.T) {
	entity := NewBaseEntity(uuid.New())

	var timestamps []time.Time
	timestamps = append(timestamps, entity.UpdatedAt())

	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Millisecond)
		entity.Update()
		timestamps = append(timestamps, entity.UpdatedAt())
	}

	for i := 1; i < len(timestamps); i++ {
		if !timestamps[i].After(timestamps[i-1]) {
			t.Errorf("Timestamp %d (%v) should be after timestamp %d (%v)", i, timestamps[i], i-1, timestamps[i-1])
		}
	}
}

func TestBaseEntity_ImmutableID(t *testing.T) {
	id := uuid.New()
	entity := NewBaseEntity(id)

	originalID := entity.ID()

	entity.Update()

	if entity.ID() != originalID {
		t.Errorf("ID should remain immutable. Original: %s, Current: %s", originalID, entity.ID())
	}
}

func TestBaseEntity_ImmutableCreatedAt(t *testing.T) {
	entity := NewBaseEntity(uuid.New())
	originalCreatedAt := entity.CreatedAt()

	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Millisecond)
		entity.Update()

		if !entity.CreatedAt().Equal(originalCreatedAt) {
			t.Errorf("CreatedAt should remain immutable. Original: %v, Current: %v", originalCreatedAt, entity.CreatedAt())
		}
	}
}
