package types

import (
	"database/sql"

	db "github.com/byeoru/kania/db/repository"
)

type GetMyRealmsRequest struct {
	UserId int64 `form:"user_id" binding:"required,number"`
}

type RealmsResponse struct {
	ID      int64         `json:"id"`
	Name    string        `json:"name"`
	OwnerID int64         `json:"owner_id"`
	Capital sql.NullInt64 `json:"capital"`
}

type GetMyRealmsResponse struct {
	APIResponse *apiResponse `json:"api_response"`
	Realms      []*RealmsResponse
}

func ToRealmsResponse(realms []db.Realm) []*RealmsResponse {
	rsRealms := []*RealmsResponse{}
	for _, r := range realms {
		realm := &RealmsResponse{
			ID:      r.ID,
			Name:    r.Name,
			OwnerID: r.OwnerID,
			Capital: r.Capital,
		}
		rsRealms = append(rsRealms, realm)
	}
	return rsRealms
}
