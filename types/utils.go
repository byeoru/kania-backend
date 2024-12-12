package types

import (
	"github.com/go-playground/validator/v10"
)

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

var ValidPoliticalEntity validator.Func = func(fl validator.FieldLevel) bool {
	if politicalEntity, ok := fl.Field().Interface().(string); ok {
		return isValidPoliticalEntity(politicalEntity)
	}
	return false
}

func isValidPoliticalEntity(politicalEntity string) bool {
	_, isExist := PoliticalEntitySet[politicalEntity]
	return isExist
}
