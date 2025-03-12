package types

import (
	"encoding/json"
)

type GetBothJsonsQueryRequest struct {
	ActionID int64 `form:"action_id"`
}

type GetBothJsonsResponse struct {
	APIResponse *apiResponse   `json:"api_response"`
	Attacker    *JsonbResponse `json:"attacker"`
	Defender    *JsonbResponse `json:"defender"`
}

type JsonbResponse struct {
	RealmID int64           `json:"realm_id"`
	Jsonb   json.RawMessage `json:"jsonb"`
}
