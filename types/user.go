package types

type SignupUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=20"`
	Nickname string `json:"nickname" binding:"required,alphanum"`
}

type SignupUserResponse struct {
	APIResponse *apiResponse `json:"api_response"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginUserResponse struct {
	APIResponse *apiResponse `json:"api_response"`
}
