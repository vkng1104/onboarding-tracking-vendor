//go:build integration

package vendor

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/database"
)

const (
	integrationCoordinatorID  = "90000000-0000-0000-0000-000000000001"
	integrationMissingActorID = "90000000-0000-0000-0000-000000000099"
)

func TestRepositoryUpdateStagePersistsBackwardTransitionAndRejectsStaleWrite(t *testing.T) {
	db := openIntegrationDatabase(t)
	vendorID := "91000000-0000-0000-0000-000000000001"
	seedIntegrationVendor(t, db, vendorID, StageKYCVerified)
	repository := NewRepository(db)
	changedAt := time.Date(2026, time.September, 13, 8, 30, 0, 0, time.UTC)

	transition, err := repository.UpdateStage(
		context.Background(),
		vendorID,
		integrationCoordinatorID,
		StageKYCVerified,
		StageKYCDocsReceived,
		changedAt,
	)
	if err != nil {
		t.Fatalf("update stage: %v", err)
	}
	if transition.PreviousStage != StageKYCVerified || transition.NewStage != StageKYCDocsReceived {
		t.Fatalf("transition = %s -> %s", transition.PreviousStage, transition.NewStage)
	}
	if !transition.OccurredAt.Equal(changedAt) {
		t.Fatalf("occurred at = %s, want %s", transition.OccurredAt, changedAt)
	}
	assertIntegrationHistoryEvent(
		t,
		db,
		vendorID,
		integrationCoordinatorID,
		StageKYCVerified,
		StageKYCDocsReceived,
		changedAt,
	)

	_, err = repository.UpdateStage(
		context.Background(),
		vendorID,
		integrationCoordinatorID,
		StageKYCVerified,
		StageActive,
		changedAt.Add(time.Minute),
	)
	if !errors.Is(err, ErrStageConflict) {
		t.Fatalf("stale update error = %v, want %v", err, ErrStageConflict)
	}

	assertIntegrationVendorState(t, db, vendorID, StageKYCDocsReceived, changedAt, 1)
}

func TestRepositoryUpdateStageRollsBackWhenHistoryInsertFails(t *testing.T) {
	db := openIntegrationDatabase(t)
	vendorID := "91000000-0000-0000-0000-000000000002"
	initialTime := time.Date(2026, time.September, 10, 8, 30, 0, 0, time.UTC)
	seedIntegrationVendorAt(t, db, vendorID, StageContractSent, initialTime)
	repository := NewRepository(db)

	_, err := repository.UpdateStage(
		context.Background(),
		vendorID,
		integrationMissingActorID,
		StageContractSent,
		StageContractSigned,
		initialTime.Add(time.Hour),
	)
	if err == nil {
		t.Fatal("update stage error = nil, want foreign-key error")
	}

	assertIntegrationVendorState(t, db, vendorID, StageContractSent, initialTime, 0)
}

func TestRepositoryUpdateStageAllowsOnlyOneConcurrentStaleWriter(t *testing.T) {
	db := openIntegrationDatabase(t)
	vendorID := "91000000-0000-0000-0000-000000000003"
	seedIntegrationVendor(t, db, vendorID, StageContractSent)
	repository := NewRepository(db)
	changedAt := time.Date(2026, time.September, 13, 9, 0, 0, 0, time.UTC)

	targets := []Stage{StageContractSigned, StageActive}
	errorsByTarget := make([]error, len(targets))
	start := make(chan struct{})
	var waitGroup sync.WaitGroup
	for index, target := range targets {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, errorsByTarget[index] = repository.UpdateStage(
				context.Background(),
				vendorID,
				integrationCoordinatorID,
				StageContractSent,
				target,
				changedAt,
			)
		}()
	}
	close(start)
	waitGroup.Wait()

	successes := 0
	conflicts := 0
	var winningStage Stage
	for index, err := range errorsByTarget {
		switch {
		case err == nil:
			successes++
			winningStage = targets[index]
		case errors.Is(err, ErrStageConflict):
			conflicts++
		default:
			t.Fatalf("concurrent update %d error = %v", index, err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes/conflicts = %d/%d, want 1/1", successes, conflicts)
	}

	assertIntegrationVendorState(t, db, vendorID, winningStage, changedAt, 1)
}

func openIntegrationDatabase(t *testing.T) *database.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is required for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	t.Cleanup(db.Close)

	if _, err := db.Exec(ctx, `
		INSERT INTO coordinators (id, name, email, password_hash)
		VALUES ($1, 'Integration Coordinator', 'integration-coordinator@demo.local', 'not-used')
		ON CONFLICT (id) DO NOTHING
	`, integrationCoordinatorID); err != nil {
		t.Fatalf("seed integration coordinator: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM coordinators WHERE id = $1`, integrationCoordinatorID)
	})

	return db
}

func seedIntegrationVendor(t *testing.T, db *database.DB, vendorID string, stage Stage) {
	t.Helper()
	seedIntegrationVendorAt(t, db, vendorID, stage, time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC))
}

func seedIntegrationVendorAt(
	t *testing.T,
	db *database.DB,
	vendorID string,
	stage Stage,
	stageEnteredAt time.Time,
) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
		INSERT INTO vendors (
			id,
			name,
			region,
			current_stage,
			stage_entered_at,
			assigned_coordinator_id
		)
		VALUES ($1, 'Integration Vendor', 'HCMC', $2, $3, $4)
	`, vendorID, stage, stageEnteredAt, integrationCoordinatorID); err != nil {
		t.Fatalf("seed integration vendor: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM vendors WHERE id = $1`, vendorID)
	})
}

func assertIntegrationVendorState(
	t *testing.T,
	db *database.DB,
	vendorID string,
	wantStage Stage,
	wantChangedAt time.Time,
	wantHistoryCount int,
) {
	t.Helper()
	var stage Stage
	var stageEnteredAt time.Time
	if err := db.QueryRow(context.Background(), `
		SELECT current_stage, stage_entered_at
		FROM vendors
		WHERE id = $1
	`, vendorID).Scan(&stage, &stageEnteredAt); err != nil {
		t.Fatalf("query vendor state: %v", err)
	}
	if stage != wantStage || !stageEnteredAt.Equal(wantChangedAt) {
		t.Fatalf("vendor state = %s at %s, want %s at %s", stage, stageEnteredAt, wantStage, wantChangedAt)
	}

	var historyCount int
	if err := db.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM vendor_stage_transitions
		WHERE vendor_id = $1
	`, vendorID).Scan(&historyCount); err != nil {
		t.Fatalf("query history count: %v", err)
	}
	if historyCount != wantHistoryCount {
		t.Fatalf("history count = %d, want %d", historyCount, wantHistoryCount)
	}
}

func assertIntegrationHistoryEvent(
	t *testing.T,
	db *database.DB,
	vendorID string,
	wantActorID string,
	wantPreviousStage Stage,
	wantNewStage Stage,
	wantChangedAt time.Time,
) {
	t.Helper()
	var actorID string
	var previousStage Stage
	var newStage Stage
	var changedAt time.Time
	if err := db.QueryRow(context.Background(), `
		SELECT changed_by_coordinator_id, previous_stage, new_stage, changed_at
		FROM vendor_stage_transitions
		WHERE vendor_id = $1
	`, vendorID).Scan(&actorID, &previousStage, &newStage, &changedAt); err != nil {
		t.Fatalf("query history event: %v", err)
	}
	if actorID != wantActorID || previousStage != wantPreviousStage || newStage != wantNewStage || !changedAt.Equal(wantChangedAt) {
		t.Fatalf(
			"history event = actor %s, %s -> %s at %s; want actor %s, %s -> %s at %s",
			actorID,
			previousStage,
			newStage,
			changedAt,
			wantActorID,
			wantPreviousStage,
			wantNewStage,
			wantChangedAt,
		)
	}
}
