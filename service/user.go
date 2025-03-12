package service

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/byeoru/kania/config"
	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/util"
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

func (s *UserService) FindUser(ctx *gin.Context, id int64) (*db.User, error) {
	return s.store.FindUserById(ctx, id)
}

// HashPassword returns the bcrypt hash of the password
func (s *UserService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func (s *UserService) Signup(ctx *gin.Context, newUser *db.CreateUserParams, nickname string) error {
	return s.store.ExecTx(ctx, func(q *db.Queries) error {
		userId, err := s.store.CreateUser(ctx, newUser)
		if err != nil {
			return err
		}

		arg := db.CreateRealmMemberParams{
			RmID:           userId,
			RealmID:        sql.NullInt64{Valid: false},
			Nickname:       nickname,
			Status:         util.None,
			PrivateCoffers: 0,
		}

		err = s.store.CreateRealmMember(ctx, &arg)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *UserService) Login(ctx *gin.Context, email string) (*db.User, error) {
	return s.store.FindUserByEmail(ctx, email)
}

// CheckPassword checks if the provided password is correct or not
func (s *UserService) CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *UserService) CreateToken(userId int64) (string, error) {
	tokenMaker := token.GetTokenMakerInstance()
	return tokenMaker.CreateToken(userId, config.GetInstance().Token.AccessTokenDuration)
}
