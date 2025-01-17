package types

import "time"

type AttackJsonRequest struct {
	OriginSector         int32     `json:"origin_sector" binding:"required,min=0"`
	TargetSector         int32     `json:"target_sector" binding:"required,min=0"`
	CurrentWorldTime     time.Time `json:"current_world_time" binding:"required"`
	ExpectedCompletionAt time.Time `json:"expected_completion_at" binding:"required"`
}

type AttackQueryRequest struct {
	LevyID int64 `form:"levy_id" binding:"required,min=0"`
}

type AttackResponse struct {
	APIResponse *apiResponse `json:"api_response"`
}

type BattlePathRequest struct {
	LevyActionID int64 `uri:"levy_action_id" binding:"required,min=0"`
}

type BattleResponse struct {
	APIResponse *apiResponse `json:"api_response"`
}
