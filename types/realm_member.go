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

// ToRealmMembers 함수
func ToRealmLevies(rows []*db.GetOurRealmLeviesRow) []*RealmLeviesResponse {
	// 중복 제거를 위한 맵
	result := make(map[string]*RealmLeviesResponse)

	// SQL 결과 처리
	for _, row := range rows {
		// 맵 키 생성 (UserID와 RealmID를 결합)
		key := fmt.Sprintf("%d:%d", row.Realm.RealmID, row.Levy.RmID)

		// Set에 없으면 ID 추가하고 초기화
		if _, exists := result[key]; !exists {
			result[key] = &RealmLeviesResponse{
				LevyAffiliation: &LevyAffiliation{
					RmID:    row.Levy.RmID,
					RealmID: row.Realm.RealmID,
				},
			}
		}

		// 레비 정보 추가
		result[key].Levies = append(result[key].Levies, ToLevyResponse(&row.Levy))
	}
	// 맵 데이터를 슬라이스로 변환
	realmMembers := make([]*RealmLeviesResponse, 0, len(result))
	for _, member := range result {
		realmMembers = append(realmMembers, member)
	}
	return realmMembers
}
