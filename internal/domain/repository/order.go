package repository

import (
	"context"

	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/google/uuid"
)

// OrderRepository domain interface
type OrderRepository interface {
	List(ctx context.Context) []entity.Order
	Get(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	Create(ctx context.Context, order *entity.Order) error
}
