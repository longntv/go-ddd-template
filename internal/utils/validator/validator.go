package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
)

var (
	// Validator is the singleton validator instance.
	Validator *validator.Validate
)

// WireSet holds the Wire providers for validator.
var WireSet = wire.NewSet(
	NewValidator,
)

// NewValidator creates a new validator instance.
func NewValidator() (*validator.Validate, error) {
	v := validator.New()
	return v, nil
}

// Validate validates the given struct.
func Validate(v interface{}) error {
	return Validator.Struct(v)
}
