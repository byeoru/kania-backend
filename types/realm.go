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

type RealmResponse struct {
	ID              int64                 `json:"id"`
	Name            string                `json:"name"`
	OwnerID         int64                 `json:"owner_id"`
	OwnerNickname   string                `json:"owner_nickname"`
	CapitalNumber   int32                 `json:"capital_number"`
	PoliticalEntity string                `json:"political_entity"`
	Color           string                `json:"color"`
	RealmCellsJson  pqtype.NullRawMessage `json:"realm_cells_json"`
}

type GetMyRealmsResponse struct {
	APIResponse *apiResponse   `json:"api_response"`
	Realm       *RealmResponse `json:"realm"`
}

type GetMeAndOthersReams struct {
	APIResponse     *apiResponse     `json:"api_response"`
	MyRealm         *RealmResponse   `json:"my_realm"`
	TheOthersRealms []*RealmResponse `json:"the_others_realms"`
}

func ToMyRealmResponse(realm db.FindRealmWithJsonRow) *RealmResponse {
	rsRealms := RealmResponse{
		ID:              realm.ID,
		Name:            realm.Name,
		OwnerID:         realm.OwnerID,
		OwnerNickname:   realm.OwnerNickname,
		CapitalNumber:   realm.CapitalNumber,
		PoliticalEntity: realm.PoliticalEntity,
		Color:           realm.Color,
		RealmCellsJson:  realm.CellsJsonb,
	}
	return &rsRealms
}

func ToTheOthersRealmsResponse(realm db.FindAllRealmsWithJsonExcludeMeRow) *RealmResponse {
	rsRealms := RealmResponse{
		ID:              realm.ID,
		Name:            realm.Name,
		OwnerID:         realm.OwnerID,
		OwnerNickname:   realm.OwnerNickname,
		CapitalNumber:   realm.CapitalNumber,
		PoliticalEntity: realm.PoliticalEntity,
		Color:           realm.Color,
		RealmCellsJson:  realm.CellsJsonb,
	}
	return &rsRealms
}

type EstablishARealmRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=10"`
	CellNumber     int32  `json:"cell_number" binding:"required"`
	ProvinceNumber int32  `json:"province_number" binding:"required"`
	RealmColor     string `json:"realm_color" binding:"required,hexColor"`
}

type EstablishARealmResponse struct {
	APIResponse *apiResponse `json:"api_response"`
	RealmId     int64        `json:"realm_id"`
}
