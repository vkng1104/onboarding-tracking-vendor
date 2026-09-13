package vendor

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

type storeStub struct {
	list    func(context.Context) ([]record, error)
	history func(context.Context, string) ([]HistoryEvent, error)
	update  func(context.Context, string, string, Stage, Stage, time.Time) (stageTransitionRecord, error)
}

func (stub storeStub) List(ctx context.Context) ([]record, error) {
	return stub.list(ctx)
}

func (stub storeStub) History(ctx context.Context, vendorID string) ([]HistoryEvent, error) {
	return stub.history(ctx, vendorID)
}

func (stub storeStub) UpdateStage(
	ctx context.Context,
	vendorID string,
	coordinatorID string,
	expectedCurrentStage Stage,
	newStage Stage,
	changedAt time.Time,
) (stageTransitionRecord, error) {
	return stub.update(ctx, vendorID, coordinatorID, expectedCurrentStage, newStage, changedAt)
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

	byID := make(map[string]Vendor, len(vendors))
	for _, vendor := range vendors {
		byID[vendor.ID] = vendor
	}

	tests := []struct {
		id        string
		wantHours int64
		wantStuck bool
		wantNext  *Stage
	}{
		{id: "on-track", wantHours: 167, wantStuck: false, wantNext: stagePointer(StageContractSigned)},
		{id: "boundary", wantHours: 168, wantStuck: false, wantNext: stagePointer(StageKYCDocsReceived)},
		{id: "stuck", wantHours: 169, wantStuck: true, wantNext: stagePointer(StageKYCVerified)},
		{id: "active", wantHours: 2400, wantStuck: false, wantNext: nil},
		{id: "future", wantHours: 0, wantStuck: false, wantNext: stagePointer(StageActive)},
	}

	for _, tt := range tests {
		vendor, ok := byID[tt.id]
		if !ok {
			t.Errorf("vendor %q missing from list", tt.id)
			continue
		}
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

func TestListPutsStuckVendorsFirstAndCompletedWorkLast(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	records := []record{
		{ID: "active", Name: "Company A", CurrentStage: StageActive, StageEnteredAt: now.Add(-100 * 24 * time.Hour)},
		{ID: "waiting-briefly", Name: "Company B", CurrentStage: StageContractSent, StageEnteredAt: now.Add(-2 * time.Hour)},
		{ID: "stuck-longest", Name: "Company C", CurrentStage: StageKYCDocsReceived, StageEnteredAt: now.Add(-40 * 24 * time.Hour)},
		{ID: "waiting-longer", Name: "Company D", CurrentStage: StageContractSigned, StageEnteredAt: now.Add(-5 * 24 * time.Hour)},
		{ID: "stuck-recently", Name: "Company E", CurrentStage: StageKYCVerified, StageEnteredAt: now.Add(-8 * 24 * time.Hour)},
		{ID: "tie-later-name", Name: "Company Z", CurrentStage: StageContractSent, StageEnteredAt: now.Add(-3 * 24 * time.Hour)},
		{ID: "tie-earlier-name", Name: "Company M", CurrentStage: StageContractSent, StageEnteredAt: now.Add(-3 * 24 * time.Hour)},
	}
	store := storeStub{list: func(context.Context) ([]record, error) { return records, nil }}
	service := newService(store, 7*24*time.Hour, func() time.Time { return now })

	vendors, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list vendors: %v", err)
	}

	got := make([]string, 0, len(vendors))
	for _, vendor := range vendors {
		got = append(got, vendor.ID)
	}
	want := []string{
		"stuck-longest",    // stuck, waiting the longest
		"stuck-recently",   // stuck, waiting less
		"waiting-longer",   // on track, but waiting the longest
		"tie-earlier-name", // equal wait, so the name decides
		"tie-later-name",
		"waiting-briefly",
		"active", // complete work sinks below anything still in progress
	}
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
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

func TestUpdateStageAllowsBackwardTransitionAndAttributesActor(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	actor := CoordinatorSummary{ID: "10000000-0000-0000-0000-000000000001", Name: "Linh Nguyen"}
	store := storeStub{
		update: func(
			_ context.Context,
			vendorID string,
			coordinatorID string,
			expectedCurrentStage Stage,
			newStage Stage,
			changedAt time.Time,
		) (stageTransitionRecord, error) {
			if vendorID != "vendor-1" {
				t.Fatalf("vendor id = %q, want vendor-1", vendorID)
			}
			if coordinatorID != actor.ID {
				t.Fatalf("coordinator id = %q, want %q", coordinatorID, actor.ID)
			}
			if expectedCurrentStage != StageKYCVerified || newStage != StageKYCDocsReceived {
				t.Fatalf("transition = %s -> %s, want KYC_VERIFIED -> KYC_DOCS_RECEIVED", expectedCurrentStage, newStage)
			}
			if !changedAt.Equal(now.UTC()) {
				t.Fatalf("changed at = %s, want %s", changedAt, now.UTC())
			}
			return stageTransitionRecord{
				ID:            "event-1",
				OccurredAt:    changedAt,
				PreviousStage: expectedCurrentStage,
				NewStage:      newStage,
			}, nil
		},
	}
	service := newService(store, 7*24*time.Hour, func() time.Time { return now })

	event, err := service.UpdateStage(
		context.Background(),
		"vendor-1",
		actor,
		StageKYCVerified,
		StageKYCDocsReceived,
	)
	if err != nil {
		t.Fatalf("update stage: %v", err)
	}
	if event.Actor != actor {
		t.Fatalf("actor = %#v, want %#v", event.Actor, actor)
	}
	if event.PreviousStage != StageKYCVerified || event.NewStage != StageKYCDocsReceived {
		t.Fatalf("event = %s -> %s", event.PreviousStage, event.NewStage)
	}
}

func TestUpdateStageRejectsInvalidAndUnchangedStagesBeforeStorage(t *testing.T) {
	store := storeStub{
		update: func(context.Context, string, string, Stage, Stage, time.Time) (stageTransitionRecord, error) {
			t.Fatal("store should not be called")
			return stageTransitionRecord{}, nil
		},
	}
	service := NewService(store, 7)
	actor := CoordinatorSummary{ID: "coordinator-1", Name: "Linh Nguyen"}

	tests := []struct {
		name     string
		expected Stage
		newStage Stage
		wantErr  error
	}{
		{name: "invalid expected stage", expected: "UNKNOWN", newStage: StageActive, wantErr: ErrInvalidStage},
		{name: "invalid new stage", expected: StageContractSent, newStage: "UNKNOWN", wantErr: ErrInvalidStage},
		{name: "unchanged stage", expected: StageContractSigned, newStage: StageContractSigned, wantErr: ErrStageUnchanged},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.UpdateStage(context.Background(), "vendor-1", actor, tt.expected, tt.newStage)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateStagePreservesConflictError(t *testing.T) {
	store := storeStub{
		update: func(context.Context, string, string, Stage, Stage, time.Time) (stageTransitionRecord, error) {
			return stageTransitionRecord{}, ErrStageConflict
		},
	}
	service := NewService(store, 7)

	_, err := service.UpdateStage(
		context.Background(),
		"vendor-1",
		CoordinatorSummary{ID: "coordinator-1"},
		StageContractSigned,
		StageActive,
	)
	if !errors.Is(err, ErrStageConflict) {
		t.Fatalf("error = %v, want %v", err, ErrStageConflict)
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
