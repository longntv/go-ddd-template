// Package crypto implements cryptographic ports such as password hashing.
package crypto

import (
	"errors"
	"fmt"

	"github.com/google/wire"
	"golang.org/x/crypto/bcrypt"

	"github.com/longntv/go-ddd-template/internal/domain/gateway"
	"github.com/longntv/go-ddd-template/internal/domain/model"
)

// WireSet holds the Wire providers for crypto.
var WireSet = wire.NewSet(
	ProvideBcryptHasher,
	wire.Bind(new(gateway.PasswordHasher), new(*BcryptHasher)),
)

// BcryptHasher implements gateway.PasswordHasher with bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a hasher with the given bcrypt cost. Tests use
// bcrypt.MinCost to stay fast; production uses bcrypt.DefaultCost.
func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

// ProvideBcryptHasher provides a hasher with bcrypt.DefaultCost.
func ProvideBcryptHasher() *BcryptHasher {
	return NewBcryptHasher(bcrypt.DefaultCost)
}

// Hash returns the bcrypt hash of plain. bcrypt only accepts up to 72 bytes;
// longer input returns model.ErrPasswordTooLong.
func (h *BcryptHasher) Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		return "", model.ErrPasswordTooLong
	}
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// Compare reports whether plain matches hash. Nothing calls it yet; it is the
// port a login use case will verify credentials through.
func (h *BcryptHasher) Compare(hash, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return model.ErrPasswordMismatch
	}
	if err != nil {
		return fmt.Errorf("compare password: %w", err)
	}
	return nil
}
