package types

import (
	"fmt"

	db "github.com/byeoru/kania/db/repository"
)

type GetRealmMembersLeviesResponse struct {
	APIResponse  *apiResponse            `json:"api_response"`
	RealmMembers []*RealmMembersResponse `json:"realm_members"`
}

type RealmMemberIDs struct {
	UserID  int64 `json:"user_id"`
	RealmID int64 `json:"realm_id"`
}

type RealmMembersResponse struct {
	RealmMember *RealmMemberIDs `json:"realm_member"`
	Levies      []*LevyResponse `json:"levies"`
}

// ToRealmMembers 함수
func ToRealmMembers(rows []*db.GetRealmMembersLeviesRow) []*RealmMembersResponse {
	// 중복 제거를 위한 맵
	result := make(map[string]*RealmMembersResponse)

	// SQL 결과 처리
	for _, row := range rows {
		// 맵 키 생성 (UserID와 RealmID를 결합)
		key := fmt.Sprintf("%d:%d", row.RealmMember.RealmID, row.RealmMember.UserID)

		// Set에 없으면 ID 추가하고 초기화
		if _, exists := result[key]; !exists {
			result[key] = &RealmMembersResponse{
				RealmMember: &RealmMemberIDs{
					UserID:  row.RealmMember.UserID,
					RealmID: row.RealmMember.RealmID,
				},
			}
		}

		// 레비 정보 추가
		result[key].Levies = append(result[key].Levies, ToLevyResponse(&row.Levy))
	}
	// 맵 데이터를 슬라이스로 변환
	realmMembers := make([]*RealmMembersResponse, 0, len(result))
	for _, member := range result {
		realmMembers = append(realmMembers, member)
	}
	return realmMembers
}
