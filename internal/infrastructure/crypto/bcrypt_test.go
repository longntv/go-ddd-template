package crypto_test

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/infrastructure/crypto"
)

func TestBcryptHasher_HashAndCompare(t *testing.T) {
	t.Parallel()

	h := crypto.NewBcryptHasher(bcrypt.MinCost)

	hash, err := h.Hash("password123")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "password123" || !strings.HasPrefix(hash, "$2a$") {
		t.Fatalf("Hash() = %q, want a bcrypt hash", hash)
	}

	tests := map[string]struct {
		hash    string
		plain   string
		wantErr bool
		wantIs  error // when set, the error must match it
		wantNot error // when set, the error must not match it
	}{
		"matching password":     {hash: hash, plain: "password123"},
		"wrong password":        {hash: hash, plain: "password124", wantErr: true, wantIs: model.ErrPasswordMismatch},
		"malformed hash":        {hash: "not-a-hash", plain: "password123", wantErr: true, wantNot: model.ErrPasswordMismatch},
		"empty password vs set": {hash: hash, plain: "", wantErr: true, wantIs: model.ErrPasswordMismatch},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := h.Compare(tt.hash, tt.plain)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Compare() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("Compare() error = %v, want %v", err, tt.wantIs)
			}
			if tt.wantNot != nil && errors.Is(err, tt.wantNot) {
				t.Errorf("Compare() error = %v, must not match %v", err, tt.wantNot)
			}
		})
	}
}

func TestBcryptHasher_HashSaltsEachCall(t *testing.T) {
	t.Parallel()

	h := crypto.NewBcryptHasher(bcrypt.MinCost)
	first, _ := h.Hash("password123")
	second, _ := h.Hash("password123")

	if first == second {
		t.Error("Hash() returned the same value twice, want a fresh salt per call")
	}
}

func TestBcryptHasher_RejectsPasswordsOver72Bytes(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"73 ASCII bytes":               strings.Repeat("a", 73),
		"40 two-byte runes (80 bytes)": strings.Repeat("é", 40),
	}
	for name, plain := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := crypto.NewBcryptHasher(bcrypt.MinCost).Hash(plain)
			if !errors.Is(err, model.ErrPasswordTooLong) {
				t.Errorf("Hash() error = %v, want %v", err, model.ErrPasswordTooLong)
			}
		})
	}
}
