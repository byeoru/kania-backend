package types

import (
	db "github.com/byeoru/kania/db/repository"
	"github.com/sqlc-dev/pqtype"
)

var PoliticalEntitySet = map[string]struct{}{
	"Tribe":                {}, // 부족
	"TribalConfederation":  {}, // 부족 연맹
	"Kingdom":              {}, // 왕국
	"KingdomConfederation": {}, // 왕국 연맹
	"Empire":               {}, // 제국
	"FeudatoryState":       {}, // 번국
}

type RealmsResponse struct {
	ID              int64                 `json:"id"`
	Name            string                `json:"name"`
	OwnerID         int64                 `json:"owner_id"`
	CapitalNumber   int32                 `json:"capital_number"`
	PoliticalEntity string                `json:"political_entity"`
	RealmCellsJson  pqtype.NullRawMessage `json:"realm_cells_json"`
}

type GetMyRealmsResponse struct {
	APIResponse *apiResponse `json:"api_response"`
	Realms      *RealmsResponse
}

func ToRealmsResponse(realm db.FindRealmWithJsonRow) *RealmsResponse {
	rsRealms := RealmsResponse{
		ID:              realm.ID,
		Name:            realm.Name,
		OwnerID:         realm.OwnerID,
		CapitalNumber:   realm.CapitalNumber,
		PoliticalEntity: realm.PoliticalEntity,
		RealmCellsJson:  realm.CellsJsonb,
	}
	return &rsRealms
}

type EstablishARealmRequest struct {
	Name            string `json:"name" binding:"required,min=1,max=10"`
	PoliticalEntity string `json:"political_entity" binding:"required,politicalEntity"`
	CellNumber      int32  `json:"cell_number" binding:"required"`
	ProvinceNumber  int32  `json:"province_number" binding:"required"`
}

type EstablishARealmResponse struct {
	APIResponse *apiResponse `json:"api_response"`
}
