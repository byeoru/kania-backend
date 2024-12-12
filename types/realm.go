package types

import (
	db "github.com/byeoru/kania/db/repository"
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
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	OwnerID         int64  `json:"owner_id"`
	CapitalNumber   int32  `json:"capital_number"`
	PoliticalEntity string `json:"political_entity"`
}

type GetMyRealmsResponse struct {
	APIResponse *apiResponse `json:"api_response"`
	Realms      []*RealmsResponse
}

func ToRealmsResponse(realms []db.Realm) []*RealmsResponse {
	rsRealms := []*RealmsResponse{}
	for _, r := range realms {
		realm := &RealmsResponse{
			ID:              r.ID,
			Name:            r.Name,
			OwnerID:         r.OwnerID,
			CapitalNumber:   r.CapitalNumber,
			PoliticalEntity: r.PoliticalEntity,
		}
		rsRealms = append(rsRealms, realm)
	}
	return rsRealms
}

type EstablishARealmRequest struct {
	Name            string `json:"name" binding:"required,min=1,max=10"`
	OwnerID         int64  `json:"owner_id" binding:"required"`
	CapitalNumber   int32  `json:"capital_number" binding:"required"`
	PoliticalEntity string `json:"political_entity" binding:"required,politicalEntity"`
	CellNumber      int32  `json:"cell_number" binding:"required"`
	ProvinceNumber  int32  `json:"province_number" binding:"required"`
}

type EstablishARealmResponse struct {
	APIResponse *apiResponse `json:"api_response"`
}
