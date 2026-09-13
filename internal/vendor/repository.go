package vendor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/database"
)

type repositoryDatabase interface {
	database.Queryer
	WithinTx(context.Context, pgx.TxOptions, func(pgx.Tx) error) error
}

type Repository struct {
	database repositoryDatabase
}

func NewRepository(database repositoryDatabase) *Repository {
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

func (r *Repository) UpdateStage(
	ctx context.Context,
	vendorID string,
	coordinatorID string,
	expectedCurrentStage Stage,
	newStage Stage,
	changedAt time.Time,
) (stageTransitionRecord, error) {
	id, err := parseVendorID(vendorID)
	if err != nil {
		return stageTransitionRecord{}, err
	}
	actorID, err := parseCoordinatorID(coordinatorID)
	if err != nil {
		return stageTransitionRecord{}, err
	}

	var transition stageTransitionRecord
	err = r.database.WithinTx(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		const lockQuery = `SELECT current_stage FROM vendors WHERE id = $1 FOR UPDATE`
		var currentStage Stage
		if err := tx.QueryRow(ctx, lockQuery, id).Scan(&currentStage); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("lock vendor: %w", err)
		}

		if currentStage != expectedCurrentStage {
			return ErrStageConflict
		}
		if currentStage == newStage {
			return ErrStageUnchanged
		}

		const updateQuery = `
			UPDATE vendors
			SET current_stage = $2, stage_entered_at = $3, updated_at = $3
			WHERE id = $1
		`
		if _, err := tx.Exec(ctx, updateQuery, id, newStage, changedAt); err != nil {
			return fmt.Errorf("update vendor: %w", err)
		}

		const insertQuery = `
			INSERT INTO vendor_stage_transitions (
				id,
				vendor_id,
				previous_stage,
				new_stage,
				changed_by_coordinator_id,
				changed_at
			)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
			RETURNING id
		`
		if err := tx.QueryRow(
			ctx,
			insertQuery,
			id,
			currentStage,
			newStage,
			actorID,
			changedAt,
		).Scan(&transition.ID); err != nil {
			return fmt.Errorf("insert vendor stage transition: %w", err)
		}

		transition.OccurredAt = changedAt
		transition.PreviousStage = currentStage
		transition.NewStage = newStage
		return nil
	})
	if err != nil {
		return stageTransitionRecord{}, err
	}

	return transition, nil
}

func parseVendorID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil || !id.Valid {
		return pgtype.UUID{}, ErrNotFound
	}
	return id, nil
}

func parseCoordinatorID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil || !id.Valid {
		return pgtype.UUID{}, fmt.Errorf("parse coordinator id")
	}
	return id, nil
}
