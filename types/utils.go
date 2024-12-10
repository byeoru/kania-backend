package types

type apiResponse struct {
	Result      bool        `json:"result"`
	Description string      `json:"description"`
	ErrorCode   interface{} `json:"errorCode"`
}

func NewAPIResponse(result bool, description string, errorCode interface{}) *apiResponse {
	return &apiResponse{
		Result:      result,
		Description: description,
		ErrorCode:   errorCode,
	}
}
