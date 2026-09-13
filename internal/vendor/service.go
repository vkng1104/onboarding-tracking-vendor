package vendor

import (
	"context"
	"fmt"
	"time"
)

type Store interface {
	List(context.Context) ([]record, error)
	History(context.Context, string) ([]HistoryEvent, error)
}

type Service struct {
	store      Store
	stuckAfter time.Duration
	now        func() time.Time
}

func NewService(store Store, stuckAfterDays int) *Service {
	return newService(store, time.Duration(stuckAfterDays)*24*time.Hour, time.Now)
}

func newService(store Store, stuckAfter time.Duration, now func() time.Time) *Service {
	return &Service{store: store, stuckAfter: stuckAfter, now: now}
}

func (s *Service) List(ctx context.Context) ([]Vendor, error) {
	records, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list vendors: %w", err)
	}

	currentTime := s.now()
	vendors := make([]Vendor, 0, len(records))
	for _, record := range records {
		elapsed := currentTime.Sub(record.StageEnteredAt)
		if elapsed < 0 {
			elapsed = 0
		}

		vendors = append(vendors, Vendor{
			ID:                  record.ID,
			Name:                record.Name,
			Region:              record.Region,
			Notes:               record.Notes,
			CurrentStage:        record.CurrentStage,
			StageEnteredAt:      record.StageEnteredAt,
			HoursInCurrentStage: int64(elapsed / time.Hour),
			IsStuck:             record.CurrentStage != StageActive && elapsed > s.stuckAfter,
			AssignedCoordinator: CoordinatorSummary{
				ID:   record.CoordinatorID,
				Name: record.CoordinatorName,
			},
			NextStage: nextStage(record.CurrentStage),
		})
	}
	return vendors, nil
}

func (s *Service) History(ctx context.Context, vendorID string) ([]HistoryEvent, error) {
	history, err := s.store.History(ctx, vendorID)
	if err != nil {
		return nil, fmt.Errorf("get vendor history: %w", err)
	}
	return history, nil
}
