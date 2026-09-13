package vendor

import "time"

type Stage string

const (
	StageContractSent    Stage = "CONTRACT_SENT"
	StageContractSigned  Stage = "CONTRACT_SIGNED"
	StageKYCDocsReceived Stage = "KYC_DOCS_RECEIVED"
	StageKYCVerified     Stage = "KYC_VERIFIED"
	StageActive          Stage = "ACTIVE"
)

type CoordinatorSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Vendor struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	Region              string             `json:"region"`
	Notes               string             `json:"notes"`
	CurrentStage        Stage              `json:"current_stage"`
	StageEnteredAt      time.Time          `json:"stage_entered_at"`
	HoursInCurrentStage int64              `json:"hours_in_current_stage"`
	IsStuck             bool               `json:"is_stuck"`
	AssignedCoordinator CoordinatorSummary `json:"assigned_coordinator"`
	NextStage           *Stage             `json:"next_stage"`
}

type HistoryEvent struct {
	ID            string             `json:"id"`
	OccurredAt    time.Time          `json:"occurred_at"`
	Actor         CoordinatorSummary `json:"actor"`
	PreviousStage Stage              `json:"previous_stage"`
	NewStage      Stage              `json:"new_stage"`
}

type record struct {
	ID              string
	Name            string
	Region          string
	Notes           string
	CurrentStage    Stage
	StageEnteredAt  time.Time
	CoordinatorID   string
	CoordinatorName string
}
