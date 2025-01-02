package types

import (
	"strconv"

	"github.com/byeoru/kania/util"
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
	_, isExist := util.PoliticalEntitySet[politicalEntity]
	return isExist
}

var ValidColor validator.Func = func(fl validator.FieldLevel) bool {
	if color, ok := fl.Field().Interface().(string); ok {
		return isValidColor(color)
	}
	return false
}

func isValidColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}

	colorHex := color[1:]
	_, err := strconv.ParseUint(colorHex, 16, 64)

	return err == nil
}

func Map[T any, R any](input []T, mapper func(T) R) []R {
	result := make([]R, len(input))
	for i, v := range input {
		result[i] = mapper(v)
	}
	return result
}
