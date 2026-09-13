package vendor

import (
	"context"
	"errors"
	"testing"
	"time"
)

type storeStub struct {
	list    func(context.Context) ([]record, error)
	history func(context.Context, string) ([]HistoryEvent, error)
}

func (stub storeStub) List(ctx context.Context) ([]record, error) {
	return stub.list(ctx)
}

func (stub storeStub) History(ctx context.Context, vendorID string) ([]HistoryEvent, error) {
	return stub.history(ctx, vendorID)
}

func TestListCalculatesTimeStuckStateAndNextStage(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	records := []record{
		{ID: "on-track", CurrentStage: StageContractSent, StageEnteredAt: now.Add(-167 * time.Hour)},
		{ID: "boundary", CurrentStage: StageContractSigned, StageEnteredAt: now.Add(-168 * time.Hour)},
		{ID: "stuck", CurrentStage: StageKYCDocsReceived, StageEnteredAt: now.Add(-169 * time.Hour)},
		{ID: "active", CurrentStage: StageActive, StageEnteredAt: now.Add(-100 * 24 * time.Hour)},
		{ID: "future", CurrentStage: StageKYCVerified, StageEnteredAt: now.Add(time.Hour)},
	}
	store := storeStub{
		list: func(context.Context) ([]record, error) { return records, nil },
		history: func(context.Context, string) ([]HistoryEvent, error) {
			t.Fatal("history should not be called")
			return nil, nil
		},
	}
	service := newService(store, 7*24*time.Hour, func() time.Time { return now })

	vendors, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list vendors: %v", err)
	}

	tests := []struct {
		index     int
		wantHours int64
		wantStuck bool
		wantNext  *Stage
	}{
		{index: 0, wantHours: 167, wantStuck: false, wantNext: stagePointer(StageContractSigned)},
		{index: 1, wantHours: 168, wantStuck: false, wantNext: stagePointer(StageKYCDocsReceived)},
		{index: 2, wantHours: 169, wantStuck: true, wantNext: stagePointer(StageKYCVerified)},
		{index: 3, wantHours: 2400, wantStuck: false, wantNext: nil},
		{index: 4, wantHours: 0, wantStuck: false, wantNext: stagePointer(StageActive)},
	}

	for _, tt := range tests {
		vendor := vendors[tt.index]
		if vendor.HoursInCurrentStage != tt.wantHours {
			t.Errorf("vendor %q hours = %d, want %d", vendor.ID, vendor.HoursInCurrentStage, tt.wantHours)
		}
		if vendor.IsStuck != tt.wantStuck {
			t.Errorf("vendor %q stuck = %t, want %t", vendor.ID, vendor.IsStuck, tt.wantStuck)
		}
		if !equalStagePointers(vendor.NextStage, tt.wantNext) {
			t.Errorf("vendor %q next stage = %v, want %v", vendor.ID, vendor.NextStage, tt.wantNext)
		}
	}
}

func TestHistoryPreservesNotFoundError(t *testing.T) {
	store := storeStub{
		list: func(context.Context) ([]record, error) { return nil, nil },
		history: func(context.Context, string) ([]HistoryEvent, error) {
			return nil, ErrNotFound
		},
	}
	service := NewService(store, 7)

	_, err := service.History(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("history error = %v, want %v", err, ErrNotFound)
	}
}

func stagePointer(stage Stage) *Stage {
	return &stage
}

func equalStagePointers(left, right *Stage) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
