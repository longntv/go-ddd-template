package gateway

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mock/password.go -source=password.go

// PasswordHasher hashes and verifies user passwords. Services only ever see
// the hash; the plain-text password must not be stored or logged.
type PasswordHasher interface {
	// Hash returns a salted one-way hash of the plain-text password. It returns
	// model.ErrPasswordTooLong when the password exceeds the algorithm's limit.
	Hash(plain string) (string, error)

	// Compare reports whether plain matches hash. It returns
	// model.ErrPasswordMismatch when it doesn't, and another error when the
	// hash is malformed.
	Compare(hash, plain string) error
}
