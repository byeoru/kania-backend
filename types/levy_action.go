package types

import (
	"time"

	db "github.com/byeoru/kania/db/repository"
)

type AttackJsonRequest struct {
	TargetSector int32 `json:"target_sector" binding:"required,min=0"`
}

type AttackQueryRequest struct {
	LevyID int64 `form:"levy_id" binding:"required,min=0"`
}

type AttackResponse struct {
	APIResponse *apiResponse        `json:"api_response"`
	LevyAction  *LevyActionResponse `json:"levy_action"`
}

type BattlePathRequest struct {
	LevyActionID int64 `uri:"levy_action_id" binding:"required,min=0"`
}

type MoveJsonRequest struct {
	TargetSector int32 `json:"target_sector" binding:"required,min=0"`
}

type MoveQueryRequest struct {
	LevyID int64 `form:"levy_id" binding:"required,min=0"`
}

type MoveResponse struct {
	APIResponse *apiResponse        `json:"api_response"`
	LevyAction  *LevyActionResponse `json:"levy_action"`
}

type LevyActionResponse struct {
	LevyActionID         int64     `json:"levy_action_id"`
	LevyID               int64     `json:"levy_id"`
	RealmID              int64     `json:"realm_id"`
	OriginSector         int32     `json:"origin_sector"`
	TargetSector         int32     `json:"target_sector"`
	ActionType           string    `json:"action_type"`
	StartedAt            time.Time `json:"started_at"`
	ExpectedCompletionAt time.Time `json:"expected_completion_at"`
}

func ToLevyActionResponse(action *db.LeviesAction) *LevyActionResponse {
	rsAction := &LevyActionResponse{
		LevyActionID:         action.LevyActionID,
		LevyID:               action.LevyID,
		RealmID:              action.RealmID,
		OriginSector:         action.OriginSector,
		TargetSector:         action.TargetSector,
		ActionType:           action.ActionType,
		StartedAt:            action.StartedAt,
		ExpectedCompletionAt: action.ExpectedCompletionAt,
	}
	return rsAction
}

type FindLevyActionQueryRequest struct {
	LevyID int64 `form:"levy_id" binding:"required,min=0"`
}

type FindLevyActionResponse struct {
	APIResponse *apiResponse        `json:"api_response"`
	LevyAction  *LevyActionResponse `json:"levy_action"`
}
