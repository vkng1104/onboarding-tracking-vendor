package vendor

import "errors"

var (
	ErrNotFound       = errors.New("vendor not found")
	ErrInvalidStage   = errors.New("invalid vendor stage")
	ErrStageConflict  = errors.New("vendor stage changed")
	ErrStageUnchanged = errors.New("vendor stage is unchanged")
)
