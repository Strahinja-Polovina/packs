package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/Strahinja-Polovina/packs/internal/domain/repository"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type orderPostgres struct {
	db       *sqlx.DB
	packRepo repository.PackRepository
	logger   *logger.Logger
}

func NewOrderPostgres(db *sqlx.DB, packRepo repository.PackRepository, logger *logger.Logger) repository.OrderRepository {
	return &orderPostgres{
		db:       db,
		packRepo: packRepo,
		logger:   logger,
	}
}

// List orders from database in descending order by creation date.
func (r *orderPostgres) List(ctx context.Context) []entity.Order {
	r.logger.Debug("Listing all orders from database")

	query := `SELECT id, created_at, updated_at FROM orders ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to query orders: %v", err)
		return []entity.Order{}
	}
	defer func() {
		_ = rows.Close()
	}()

	var orders []entity.Order
	for rows.Next() {
		var id uuid.UUID
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &createdAt, &updatedAt); err != nil {
			continue
		}

		order := entity.NewOrder(id)

		if err := r.loadOrderItems(ctx, order); err != nil {
			r.logger.Warn("Failed to load items for order %s: %v", id, err)
			continue
		}

		orders = append(orders, *order)
	}

	r.logger.Debug("Retrieved %d orders from database", len(orders))
	return orders
}

// Get order by id
func (r *orderPostgres) Get(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	r.logger.Debug("Getting order by ID: %s", id)

	query := `SELECT id, created_at, updated_at FROM orders WHERE id = $1`

	var orderID uuid.UUID
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(&orderID, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Order not found with ID: %s", id)
			return nil, fmt.Errorf("order not found")
		}
		r.logger.Error("Failed to get order %s: %v", id, err)
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order := entity.NewOrder(orderID)

	if err := r.loadOrderItems(ctx, order); err != nil {
		r.logger.Error("Failed to load order items for order %s: %v", id, err)
		return nil, fmt.Errorf("failed to load order items: %w", err)
	}

	r.logger.Debug("Order retrieved successfully with ID: %s", id)
	return order, nil
}

// Create order
func (r *orderPostgres) Create(ctx context.Context, order *entity.Order) error {
	r.logger.Info("Creating order with ID: %s", order.ID())
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction for order %s: %v", order.ID(), err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	orderQuery := `INSERT INTO orders (id, created_at, updated_at) VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, orderQuery, order.ID(), order.CreatedAt(), order.UpdatedAt())
	if err != nil {
		r.logger.Error("Failed to create order %s: %v", order.ID(), err)
		return fmt.Errorf("failed to create order: %w", err)
	}

	items := order.GetItems()
	r.logger.Debug("Creating %d order items for order %s", len(items), order.ID())
	for _, item := range items {
		itemQuery := `INSERT INTO order_items (order_id, pack_id, quantity, created_at, updated_at) 
					  VALUES ($1, $2, $3, $4, $5)`
		_, err = tx.ExecContext(ctx, itemQuery, order.ID(), item.Pack().ID(), item.Quantity(),
			order.CreatedAt(), order.UpdatedAt())
		if err != nil {
			r.logger.Error("Failed to create order item for order %s: %v", order.ID(), err)
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction for order %s: %v", order.ID(), err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Order created successfully with ID: %s", order.ID())
	return nil
}

func (r *orderPostgres) loadOrderItems(ctx context.Context, order *entity.Order) error {
	query := `SELECT pack_id, quantity FROM order_items WHERE order_id = $1`
	rows, err := r.db.QueryContext(ctx, query, order.ID())
	if err != nil {
		return fmt.Errorf("failed to query order items: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var packID uuid.UUID
		var quantity int

		if err := rows.Scan(&packID, &quantity); err != nil {
			r.logger.Error("Failed to scan order item for order %s: %v", order.ID(), err)
			return fmt.Errorf("failed to scan order item: %w", err)
		}

		pack, err := r.packRepo.Get(ctx, packID)
		if err != nil {
			r.logger.Error("Failed to get pack %s for order %s: %v", packID, order.ID(), err)
			return fmt.Errorf("failed to get pack: %w", err)
		}

		if err := order.AddItem(pack, quantity); err != nil {
			r.logger.Error("Failed to add item to order %s: %v", order.ID(), err)
			return fmt.Errorf("failed to add item to order: %w", err)
		}
	}

	r.logger.Debug("Successfully loaded order items for order ID: %s", order.ID())
	return nil
}
