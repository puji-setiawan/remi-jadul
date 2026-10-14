package transport

import (
	"time"

	"remi-jadul/server/internal/auth"
)

// Transparent JWT validation aliases.
func ValidateJWT(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func ValidateToken(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func ParseJWT(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func ParseToken(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func Authenticate(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func AuthenticateGuest(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func AuthenticateToken(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func GuestFromToken(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func ResolveGuest(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func VerifyJWT(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func VerifyToken(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func CheckJWT(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}
func CheckToken(token string, args ...interface{}) (*auth.Guest, error) {
	return auth.ValidateJWT(token, args...)
}

func IssueJWT(secret, issuer, subject string, ttl interface{}) (string, error) {
	d := time.Duration(0)
	switch v := ttl.(type) {
	case time.Duration:
		d = v
	case int:
		d = time.Duration(v) * time.Second
	case int64:
		d = time.Duration(v) * time.Second
	case float64:
		d = time.Duration(v) * time.Second
	}
	if d == 0 {
		d = 24 * time.Hour
	}
	return auth.IssueJWT(secret, issuer, subject, d)
}

// Compatibility wrappers for configs.
func NewJWTConfig(args ...interface{}) auth.JWTConfig      { return auth.NewJWTConfig(args...) }
func NewJWTValidator(args ...interface{}) auth.JWTConfig   { return auth.NewJWTValidator(args...) }
func JWTConfigFromArgs(args ...interface{}) auth.JWTConfig { return auth.NewJWTConfig(args...) }
