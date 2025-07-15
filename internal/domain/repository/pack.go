package repository

import (
	"context"

	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/google/uuid"
)

// PackRepository domain interface
type PackRepository interface {
	List(ctx context.Context) []entity.Pack
	Get(ctx context.Context, id uuid.UUID) (*entity.Pack, error)
	Create(ctx context.Context, pack *entity.Pack) error
	Update(ctx context.Context, pack *entity.Pack) error
	Delete(ctx context.Context, pack *entity.Pack) error
	ExistsBySize(ctx context.Context, size int) (bool, error)
}
