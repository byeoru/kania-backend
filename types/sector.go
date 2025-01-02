package types

type GetPopulationRequest struct {
	CellNumber int32 `uri:"cell_number" binding:"required,min=1"`
}

type GetPopulationResponse struct {
	APIResponse *apiResponse `json:"api_response"`
	Population  int32        `json:"population"`
}
