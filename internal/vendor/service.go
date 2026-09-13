package vendor

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"
)

type store interface {
	List(context.Context) ([]record, error)
	History(context.Context, string) ([]HistoryEvent, error)
	UpdateStage(context.Context, string, string, Stage, Stage, time.Time) (stageTransitionRecord, error)
}

type Service struct {
	store      store
	stuckAfter time.Duration
	now        func() time.Time
}

func NewService(store store, stuckAfterDays int) *Service {
	return newService(store, time.Duration(stuckAfterDays)*24*time.Hour, time.Now)
}

func newService(store store, stuckAfter time.Duration, now func() time.Time) *Service {
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

	slices.SortFunc(vendors, byAttentionNeeded)
	return vendors, nil
}

func (s *Service) History(ctx context.Context, vendorID string) ([]HistoryEvent, error) {
	history, err := s.store.History(ctx, vendorID)
	if err != nil {
		return nil, fmt.Errorf("get vendor history: %w", err)
	}
	return history, nil
}

func (s *Service) UpdateStage(
	ctx context.Context,
	vendorID string,
	actor CoordinatorSummary,
	expectedCurrentStage Stage,
	newStage Stage,
) (HistoryEvent, error) {
	if !isValidStage(expectedCurrentStage) || !isValidStage(newStage) {
		return HistoryEvent{}, ErrInvalidStage
	}
	if expectedCurrentStage == newStage {
		return HistoryEvent{}, ErrStageUnchanged
	}

	transition, err := s.store.UpdateStage(
		ctx,
		vendorID,
		actor.ID,
		expectedCurrentStage,
		newStage,
		s.now().UTC(),
	)
	if err != nil {
		return HistoryEvent{}, fmt.Errorf("update vendor stage: %w", err)
	}

	return HistoryEvent{
		ID:            transition.ID,
		OccurredAt:    transition.OccurredAt,
		Actor:         actor,
		PreviousStage: transition.PreviousStage,
		NewStage:      transition.NewStage,
	}, nil
}

// attentionRank ranks a vendor by how much coordinator attention it needs: stuck
// vendors first, then vendors still moving through the workflow, then completed ones.
func attentionRank(vendor Vendor) int {
	switch {
	case vendor.IsStuck:
		return 0
	case vendor.CurrentStage == StageActive:
		return 2
	default:
		return 1
	}
}

// byAttentionNeeded puts the most urgent vendor first so a coordinator scanning the
// dashboard sees stalled work before anything else. Ties fall back to the longest
// wait and then to the name, so the order is stable for the same data.
func byAttentionNeeded(left, right Vendor) int {
	if rank := cmp.Compare(attentionRank(left), attentionRank(right)); rank != 0 {
		return rank
	}
	if wait := cmp.Compare(right.HoursInCurrentStage, left.HoursInCurrentStage); wait != 0 {
		return wait
	}
	return cmp.Compare(left.Name, right.Name)
}
