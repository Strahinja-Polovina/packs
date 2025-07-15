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

type packPostgres struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func NewPackPostgres(db *sqlx.DB, logger *logger.Logger) repository.PackRepository {
	return &packPostgres{
		db:     db,
		logger: logger,
	}
}

// List packs from database in ascending order by size.
func (r *packPostgres) List(ctx context.Context) []entity.Pack {
	r.logger.Debug("Listing all packs from database")
	query := `SELECT id, size, created_at, updated_at FROM packs ORDER BY size`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to query packs: %v", err)
		return []entity.Pack{}
	}
	defer func() {
		_ = rows.Close()
	}()

	var packs []entity.Pack
	for rows.Next() {
		var id uuid.UUID
		var size int
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &size, &createdAt, &updatedAt); err != nil {
			continue
		}

		pack, err := entity.NewPack(id, size)
		if err != nil {
			continue
		}

		packs = append(packs, *pack)
	}

	return packs
}

// Get pack by id
func (r *packPostgres) Get(ctx context.Context, id uuid.UUID) (*entity.Pack, error) {
	r.logger.Debug("Getting pack by ID: %s", id)
	query := `SELECT id, size, created_at, updated_at FROM packs WHERE id = $1`

	var packID uuid.UUID
	var size int
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(&packID, &size, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Pack not found with ID: %s", id)
			return nil, fmt.Errorf("pack not found")
		}
		r.logger.Error("Failed to get pack %s: %v", id, err)
		return nil, fmt.Errorf("failed to get pack: %w", err)
	}

	pack, err := entity.NewPack(packID, size)
	if err != nil {
		return nil, fmt.Errorf("failed to create pack entity: %w", err)
	}

	return pack, nil
}

// Create pack
func (r *packPostgres) Create(ctx context.Context, pack *entity.Pack) error {
	r.logger.Info("Creating pack with ID: %s, size: %d", pack.ID(), pack.Size())
	query := `INSERT INTO packs (id, size, created_at, updated_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, pack.ID(), pack.Size(), pack.CreatedAt(), pack.UpdatedAt())
	if err != nil {
		r.logger.Error("Failed to create pack %s: %v", pack.ID(), err)
		return fmt.Errorf("failed to create pack: %w", err)
	}

	r.logger.Info("Pack created successfully with ID: %s", pack.ID())
	return nil
}

// Update pack
func (r *packPostgres) Update(ctx context.Context, pack *entity.Pack) error {
	r.logger.Info("Updating pack with ID: %s, new size: %d", pack.ID(), pack.Size())
	query := `UPDATE packs SET size = $2, updated_at = $3 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, pack.ID(), pack.Size(), pack.UpdatedAt())
	if err != nil {
		r.logger.Error("Failed to update pack %s: %v", pack.ID(), err)
		return fmt.Errorf("failed to update pack: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected for pack %s: %v", pack.ID(), err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("Pack not found for update with ID: %s", pack.ID())
		return fmt.Errorf("pack not found")
	}

	r.logger.Info("Pack updated successfully with ID: %s", pack.ID())
	return nil
}

// Delete pack
func (r *packPostgres) Delete(ctx context.Context, pack *entity.Pack) error {
	r.logger.Info("Deleting pack with ID: %s", pack.ID())
	query := `DELETE FROM packs WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, pack.ID())
	if err != nil {
		r.logger.Error("Failed to delete pack %s: %v", pack.ID(), err)
		return fmt.Errorf("failed to delete pack: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected for pack deletion %s: %v", pack.ID(), err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("Pack not found for deletion with ID: %s", pack.ID())
		return fmt.Errorf("pack not found")
	}

	r.logger.Info("Pack deleted successfully with ID: %s", pack.ID())
	return nil
}

// ExistsBySize check is pack exists
func (r *packPostgres) ExistsBySize(ctx context.Context, size int) (bool, error) {
	r.logger.Debug("Checking if pack size exists: %d", size)
	query := `SELECT EXISTS(SELECT 1 FROM packs WHERE size = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, size).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check if pack size %d exists: %v", size, err)
		return false, fmt.Errorf("failed to check pack size existence: %w", err)
	}

	r.logger.Debug("Pack size %d exists: %t", size, exists)
	return exists, nil
}
