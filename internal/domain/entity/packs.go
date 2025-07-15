package entity

import (
	"github.com/google/uuid"
)

type Pack struct {
	BaseEntity
	size int
}

func NewPack(id uuid.UUID, size int) (*Pack, error) {
	if size <= 0 {
		return nil, ErrPackSize
	}

	return &Pack{
		BaseEntity: NewBaseEntity(id),
		size:       size,
	}, nil
}

func (p *Pack) Size() int {
	return p.size
}

func (p *Pack) ChangeSize(size int) error {
	if size <= 0 {
		return ErrPackSize
	}

	p.size = size
	p.Update()

	return nil
}
