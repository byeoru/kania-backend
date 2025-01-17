package types

import db "github.com/byeoru/kania/db/repository"

type CreateLevyRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=15"`
	Encampment    int32  `json:"encampment" binding:"required"`
	Swordmen      int32  `json:"swordmen" validate:"exists,min=0"`
	ShieldBearers int32  `json:"shield_bearers" validate:"exists,min=0"`
	Archers       int32  `json:"archers" validate:"exists,min=0"`
	Lancers       int32  `json:"lancers" validate:"exists,min=0"`
	SupplyTroop   int32  `json:"supply_troop" validate:"exists,min=0"`
}

type CreateLevyResponse struct {
	APIResponse     *apiResponse     `json:"api_response"`
	StateCoffers    int32            `json:"state_coffers"`
	Population      int32            `json:"population"`
	Levy            *LevyResponse    `json:"levy"`
	LevyAffiliation *LevyAffiliation `json:"levy_affiliation"`
}

func ToLevyResponse(levy *db.Levy) *LevyResponse {
	rsLevy := LevyResponse{
		LevyID:        levy.LevyID,
		Name:          levy.Name,
		Encampment:    levy.Encampment,
		Swordmen:      levy.Swordmen,
		ShieldBearers: levy.ShieldBearers,
		Archers:       levy.Archers,
		Lancers:       levy.Lancers,
		SupplyTroop:   levy.SupplyTroop,
		MovementSpeed: levy.MovementSpeed,
	}
	return &rsLevy
}

type LevyResponse struct {
	LevyID        int64   `json:"levy_id"`
	Name          string  `json:"name"`
	Encampment    int32   `json:"encampment"`
	Swordmen      int32   `json:"swordmen"`
	ShieldBearers int32   `json:"shield_bearers"`
	Archers       int32   `json:"archers"`
	Lancers       int32   `json:"lancers"`
	SupplyTroop   int32   `json:"supply_troop"`
	MovementSpeed float64 `json:"movement_speed"`
}

type CreateLevyResultInfo struct {
	StateCoffers int32
	Population   int32
}
