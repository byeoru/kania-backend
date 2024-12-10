package token

import (
	"fmt"
	"sync"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	pasetoMakerInit     sync.Once
	pasetoMakerInstance Maker
)

// PasetoMaker is a PASETO token maker
type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func InitPasetoMaker(symmetricKey string) error {
	var err error
	pasetoMakerInit.Do(func() {
		if len(symmetricKey) != chacha20poly1305.KeySize {
			err = fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
		}

		pasetoMakerInstance = &PasetoMaker{
			paseto:       paseto.NewV2(),
			symmetricKey: []byte(symmetricKey),
		}
	})
	return err
}

func GetTokenMakerInstance() Maker {
	return pasetoMakerInstance
}

// CreateToken implements Maker.
func (maker *PasetoMaker) CreateToken(userId int64, duration time.Duration) (string, error) {
	payload, err := NewPayload(userId, duration)
	if err != nil {
		return "", err
	}

	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

// VerifyToken implements Maker.
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
