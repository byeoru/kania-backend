package service

import (
	"fmt"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	userServiceInit     sync.Once
	userServiceInstance *UserService
)

type UserService struct {
	store db.Store
}

func newUserService(store db.Store) *UserService {
	userServiceInit.Do(func() {
		userServiceInstance = &UserService{
			store,
		}
	})
	return userServiceInstance
}

// HashPassword returns the bcrypt hash of the password
func (s *UserService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func (s *UserService) CreateUser(ctx *gin.Context, newUser *db.CreateUserParams) error {
	return s.store.CreateUser(ctx, *newUser)
}
