package types

import (
	"time"

	db "github.com/byeoru/kania/db/repository"
	"github.com/sqlc-dev/pqtype"
)

type RealmResponse struct {
	ID              int64                 `json:"id"`
	Name            string                `json:"name"`
	OwnerNickname   string                `json:"owner_nickname"`
	Capitals        []int32               `json:"capitals"`
	PoliticalEntity string                `json:"political_entity"`
	Color           string                `json:"color"`
	RealmCellsJson  pqtype.NullRawMessage `json:"realm_cells_json"`
}

type MyRealmResponse struct {
	*RealmResponse
	PopulationGrowthRate float64   `json:"population_growth_rate"`
	StateCoffers         int32     `json:"state_coffers"`
	CensusAt             time.Time `json:"census_at"`
	TaxCollectionAt      time.Time `json:"tax_collection_at"`
}

type GetMyRealmsResponse struct {
	APIResponse *apiResponse     `json:"api_response"`
	Realm       *MyRealmResponse `json:"realm"`
}

type GetMeAndOthersReams struct {
	APIResponse     *apiResponse     `json:"api_response"`
	MyRealm         *MyRealmResponse `json:"my_realm"`
	TheOthersRealms []*RealmResponse `json:"the_others_realms"`
}

func ToMyRealmResponse(realm *db.FindRealmWithJsonRow) *MyRealmResponse {
	if realm == nil {
		return nil
	}
	rsRealms := MyRealmResponse{
		RealmResponse: &RealmResponse{
			ID:              realm.RealmID,
			Name:            realm.Name,
			OwnerNickname:   realm.OwnerNickname,
			Capitals:        realm.Capitals,
			PoliticalEntity: realm.PoliticalEntity,
			Color:           realm.Color,
			RealmCellsJson:  realm.CellsJsonb,
		},
		PopulationGrowthRate: realm.PopulationGrowthRate,
		StateCoffers:         realm.StateCoffers,
		CensusAt:             realm.CensusAt,
		TaxCollectionAt:      realm.TaxCollectionAt,
	}
	return &rsRealms
}

func ToMyRealmFromEntityResponse(realm *db.Realm) *MyRealmResponse {
	rsRealms := MyRealmResponse{
		RealmResponse: &RealmResponse{
			ID:              realm.RealmID,
			Name:            realm.Name,
			OwnerNickname:   realm.OwnerNickname,
			Capitals:        realm.Capitals,
			PoliticalEntity: realm.PoliticalEntity,
			Color:           realm.Color,
		},
		PopulationGrowthRate: realm.PopulationGrowthRate,
		StateCoffers:         realm.StateCoffers,
		CensusAt:             realm.CensusAt,
		TaxCollectionAt:      realm.TaxCollectionAt,
	}
	return &rsRealms
}

func ToTheOthersRealmsResponse(realm *db.FindAllRealmsWithJsonExcludeMeRow) *RealmResponse {
	rsRealms := RealmResponse{
		ID:              realm.RealmID,
		Name:            realm.Name,
		OwnerNickname:   realm.OwnerNickname,
		Capitals:        realm.Capitals,
		PoliticalEntity: realm.PoliticalEntity,
		Color:           realm.Color,
		RealmCellsJson:  realm.CellsJsonb,
	}
	return &rsRealms
}

type EstablishARealmRequest struct {
	Name           string    `json:"name" binding:"required,min=1,max=10"`
	CellNumber     int32     `json:"cell_number" binding:"required"`
	ProvinceNumber int32     `json:"province_number" binding:"required"`
	RealmColor     string    `json:"realm_color" binding:"required,hexColor"`
	InitDate       time.Time `json:"init_date" binding:"required"`
	Population     int32     `json:"population" binding:"required,min=0"`
}

type EstablishARealmResponse struct {
	APIResponse *apiResponse     `json:"api_response"`
	MyRealm     *MyRealmResponse `json:"my_realm"`
}

type ExecuteCensusRequest struct {
	CurrentDate time.Time `json:"current_date" binding:"required"`
}

type ExecuteCensusResponse struct {
	APIResponse *apiResponse `json:"api_response"`
}
