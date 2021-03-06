package extension

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"math/big"
	"strconv"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"

	"golang.org/x/crypto/pbkdf2"
)

// PINExtension ...
type PINExtension interface {
	EncryptPIN(rawPwd string, options *Options) (string, string)
	ComparePIN(rawPwd string, salt string, encodedPwd string, options *Options) bool
	GenerateTempPIN(ctx context.Context) (string, error)
}

// PINExtensionImpl ...
type PINExtensionImpl struct {
}

const (
	// DefaultSaltLen is the length of generated salt for the user is 256
	DefaultSaltLen = 256
	// defaultIterations is the iteration count in PBKDF2 function is 10000
	defaultIterations = 10000
	// DefaultKeyLen is the length of encoded key in PBKDF2 function is 512
	DefaultKeyLen = 512
	// alphanumeric character used for generation of a `salt`
	alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// Default min length of the pin
	minPinLength = 4
	// Default max length of the pin
	maxPinLength = 6
	// Default length of a generated pin
	generatedPinLength = 4
)

// DefaultHashFunction ...
var DefaultHashFunction = sha512.New

// Options is a struct for custom values of salt length, number of iterations, the encoded key's length,
// and the hash function being used. If set to `nil`, default options are used:
// &Options{ 256, 10000, 512, "sha512" }
type Options struct {
	SaltLen      int
	Iterations   int
	KeyLen       int
	HashFunction func() hash.Hash
}

// NewPINExtensionImpl ...
func NewPINExtensionImpl() PINExtension {
	return &PINExtensionImpl{}
}

func generateSalt(length int) []byte {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil
	}
	for key, val := range salt {
		salt[key] = alphanum[val%byte(len(alphanum))]
	}
	return salt
}

// EncryptPIN takes two arguments, a raw pin, and a pointer to an Options struct.
// In order to use default options, pass `nil` as the second argument.
// It returns the generated salt and encoded key for the user.
func (p PINExtensionImpl) EncryptPIN(rawPwd string, options *Options) (string, string) {
	if options == nil {
		salt := generateSalt(DefaultSaltLen)
		encodedPwd := pbkdf2.Key([]byte(rawPwd), salt, defaultIterations, DefaultKeyLen, DefaultHashFunction)
		return string(salt), hex.EncodeToString(encodedPwd)
	}
	salt := generateSalt(options.SaltLen)
	encodedPwd := pbkdf2.Key([]byte(rawPwd), salt, options.Iterations, options.KeyLen, options.HashFunction)
	return string(salt), hex.EncodeToString(encodedPwd)
}

// ComparePIN takes four arguments, the raw password, its generated salt, the encoded password,
// and a pointer to the Options struct, and returns a boolean value determining whether the password is the correct one or not.
// Passing `nil` as the last argument resorts to default options.
func (p PINExtensionImpl) ComparePIN(rawPwd string, salt string, encodedPwd string, options *Options) bool {
	if options == nil {
		return encodedPwd == hex.EncodeToString(pbkdf2.Key([]byte(rawPwd), []byte(salt), defaultIterations, DefaultKeyLen, DefaultHashFunction))
	}
	return encodedPwd == hex.EncodeToString(pbkdf2.Key([]byte(rawPwd), []byte(salt), options.Iterations, options.KeyLen, options.HashFunction))
}

// GenerateTempPIN generates a temporary One Time PIN for a user
// The PIN will have 4 digits formatted as a string
func (p PINExtensionImpl) GenerateTempPIN(ctx context.Context) (string, error) {
	var pin string

	length := 0
	for length < generatedPinLength {
		number, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}

		pin += number.String()

		length++
	}

	return pin, nil
}

// ValidatePINDigits validates user pin to ensure a PIN only contains digits
func ValidatePINDigits(pin string) error {
	// ensure pin is only digits
	_, err := strconv.ParseUint(pin, 10, 64)
	if err != nil {
		return exceptions.ValidatePINDigitsError(err)
	}
	return nil
}

// ValidatePINLength validates user pin to ensure it is
// 4,5, or six digits
func ValidatePINLength(pin string) error {
	// make sure pin length is [4-6]
	if len(pin) < minPinLength || len(pin) > maxPinLength {
		return exceptions.ValidatePINLengthError(fmt.Errorf("pin should be of 4,5, or six digits"))
	}
	return nil
}
