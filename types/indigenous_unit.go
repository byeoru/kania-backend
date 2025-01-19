package types

import db "github.com/byeoru/kania/db/repository"

type GetIndigenousUnitPathRequest struct {
	SectorNumber int32 `uri:"sector_number" binding:"required,min=0"`
}

type GetIndigenousUnitResponse struct {
	APIResponse    *apiResponse            `json:"api_response"`
	IndigenousUnit *indigenousUnitResponse `json:"indigenous_unit"`
}

type indigenousUnitResponse struct {
	SectorNumber      int32 `json:"sector_number"`
	Swordmen          int32 `json:"swordmen"`
	Archers           int32 `json:"archers"`
	Lancers           int32 `json:"lancers"`
	OffensiveStrength int32 `json:"offensive_strength"`
	DefensiveStrength int32 `json:"defensive_strength"`
}

func NewIndigenousUnitResponse(entity *db.IndigenousUnit) *indigenousUnitResponse {
	return &indigenousUnitResponse{
		SectorNumber: entity.SectorNumber,
		Swordmen:     entity.Swordmen,
		Archers:      entity.Archers,
		Lancers:      entity.Lancers,
	}
}
