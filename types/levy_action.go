package types

type AttackJsonRequest struct {
	TargetSector int32 `json:"target_sector" binding:"required,min=0"`
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

type MoveJsonRequest struct {
	TargetSector int32 `json:"target_sector" binding:"required,min=0"`
}

type MoveQueryRequest struct {
	LevyID int64 `form:"levy_id" binding:"required,min=0"`
}

type MoveResponse struct {
	APIResponse *apiResponse `json:"api_response"`
}
