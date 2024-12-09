package types

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=20"`
	Nickname string `json:"nickname" binding:"required,alphanum"`
}
