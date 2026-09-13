package vendor

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/database"
)

type Repository struct {
	database database.Queryer
}

func NewRepository(database database.Queryer) *Repository {
	return &Repository{database: database}
}

func (r *Repository) List(ctx context.Context) ([]record, error) {
	const query = `
		SELECT
			v.id,
			v.name,
			v.region,
			v.notes,
			v.current_stage,
			v.stage_entered_at,
			c.id,
			c.name
		FROM vendors v
		JOIN coordinators c ON c.id = v.assigned_coordinator_id
		ORDER BY v.name
	`

	rows, err := r.database.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query vendors: %w", err)
	}
	defer rows.Close()

	vendors := make([]record, 0)
	for rows.Next() {
		var vendor record
		if err := rows.Scan(
			&vendor.ID,
			&vendor.Name,
			&vendor.Region,
			&vendor.Notes,
			&vendor.CurrentStage,
			&vendor.StageEnteredAt,
			&vendor.CoordinatorID,
			&vendor.CoordinatorName,
		); err != nil {
			return nil, fmt.Errorf("scan vendor: %w", err)
		}
		vendors = append(vendors, vendor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vendors: %w", err)
	}

	return vendors, nil
}

func (r *Repository) History(ctx context.Context, vendorID string) ([]HistoryEvent, error) {
	id, err := parseVendorID(vendorID)
	if err != nil {
		return nil, err
	}

	const existsQuery = `SELECT EXISTS (SELECT 1 FROM vendors WHERE id = $1)`
	var exists bool
	if err := r.database.QueryRow(ctx, existsQuery, id).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check vendor exists: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	const query = `
		SELECT
			t.id,
			t.changed_at,
			c.id,
			c.name,
			t.previous_stage,
			t.new_stage
		FROM vendor_stage_transitions t
		JOIN coordinators c ON c.id = t.changed_by_coordinator_id
		WHERE t.vendor_id = $1
		ORDER BY t.changed_at DESC, t.id DESC
	`

	rows, err := r.database.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query vendor history: %w", err)
	}
	defer rows.Close()

	history := make([]HistoryEvent, 0)
	for rows.Next() {
		var event HistoryEvent
		if err := rows.Scan(
			&event.ID,
			&event.OccurredAt,
			&event.Actor.ID,
			&event.Actor.Name,
			&event.PreviousStage,
			&event.NewStage,
		); err != nil {
			return nil, fmt.Errorf("scan vendor history: %w", err)
		}
		history = append(history, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vendor history: %w", err)
	}

	return history, nil
}

func parseVendorID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil || !id.Valid {
		return pgtype.UUID{}, ErrNotFound
	}
	return id, nil
}
