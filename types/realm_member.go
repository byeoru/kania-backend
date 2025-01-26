package types

import (
	"fmt"

	db "github.com/byeoru/kania/db/repository"
)

type GetRealmMembersLeviesResponse struct {
	APIResponse *apiResponse           `json:"api_response"`
	RealmLevies []*RealmLeviesResponse `json:"realm_levies"`
}

type LevyAffiliation struct {
	RmID    int64 `json:"rm_id"`
	RealmID int64 `json:"realm_id"`
}

type RealmLeviesResponse struct {
	LevyAffiliation *LevyAffiliation `json:"levy_affiliation"`
	Levies          []*LevyResponse  `json:"levies"`
}

func ToRealmLevies(rows []*db.Levy) []*RealmLeviesResponse {
	// 중복 제거를 위한 맵
	result := make(map[string]*RealmLeviesResponse)

	// SQL 결과 처리
	for _, row := range rows {
		// 멸망한 국가 부대는 제외
		if !row.RealmID.Valid {
			continue
		}
		// 맵 키 생성 (UserID와 RealmID를 결합)
		key := fmt.Sprintf("%d:%d", row.RealmID.Int64, row.RmID)

		// Set에 없으면 ID 추가하고 초기화
		if _, exists := result[key]; !exists {
			result[key] = &RealmLeviesResponse{
				LevyAffiliation: &LevyAffiliation{
					RmID:    row.RmID,
					RealmID: row.RealmID.Int64,
				},
			}
		}

		// 레비 정보 추가
		result[key].Levies = append(result[key].Levies, ToLevyResponse(row))
	}
	// 맵 데이터를 슬라이스로 변환
	resultLevies := make([]*RealmLeviesResponse, 0, len(result))
	for _, member := range result {
		resultLevies = append(resultLevies, member)
	}
	return resultLevies
}

type GetMyRealmsResponse struct {
	APIResponse *apiResponse     `json:"api_response"`
	Realm       *MyRealmResponse `json:"realm"`
}

type GetMeAndOthersReamsResponse struct {
	APIResponse     *apiResponse     `json:"api_response"`
	StandardTimes   *StandardTimes   `json:"standard_times"`
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
