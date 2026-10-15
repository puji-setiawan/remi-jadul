package transport

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
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

func validateTokenWithRepo(token, secret, issuer string, repo GuestIdentityStore) (*auth.Guest, error) {
	if repo == nil {
		return nil, auth.ErrInvalidToken
	}
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return nil, auth.ErrMalformedToken
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, auth.ErrMalformedToken
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(sigBytes) == 0 {
		return nil, auth.ErrInvalidSignature
	}
	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, auth.ErrMalformedToken
	}
	if alg, _ := header["alg"].(string); alg != "HS256" {
		return nil, auth.ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sigBytes, mac.Sum(nil)) {
		return nil, auth.ErrInvalidSignature
	}
	claims, err := auth.ParseTokenClaims(token)
	if err != nil || claims == nil {
		return nil, auth.ErrInvalidClaims
	}
	if issuer != "" && claims.Issuer != issuer {
		return nil, auth.ErrWrongIssuer
	}
	if claims.Expiry <= time.Now().Unix() {
		return nil, auth.ErrExpiredToken
	}
	if claims.Subject == "" {
		return nil, auth.ErrGuestNotFound
	}
	return repo.GetGuest(claims.Subject)
}
