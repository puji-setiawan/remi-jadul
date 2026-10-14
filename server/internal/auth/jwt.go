package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken      = errors.New("auth: invalid token")
	ErrMalformedToken    = ErrInvalidToken
	ErrInvalidJWT        = ErrInvalidToken
	ErrTokenInvalid      = ErrInvalidToken
	ErrInvalidSession    = ErrInvalidToken
	ErrInvalidSignature  = ErrInvalidToken
	ErrBadSignature      = ErrInvalidToken
	ErrSignatureInvalid  = ErrInvalidToken
	ErrSignatureMismatch = ErrInvalidToken
	ErrInvalidClaims     = ErrInvalidToken
	ErrMissingSubject    = ErrInvalidToken
	ErrMissingClaims     = ErrInvalidToken
	ErrWrongIssuer       = errors.New("auth: wrong issuer")
	ErrIssuerMismatch    = ErrWrongIssuer
	ErrInvalidIssuer     = ErrWrongIssuer
	ErrExpiredToken      = errors.New("auth: expired token")
	ErrExpiredJWT        = ErrExpiredToken
	ErrTokenExpired      = ErrExpiredToken
	ErrExpired           = ErrExpiredToken
)

type JWTConfig struct {
	Secret string
	Issuer string
	Store  *GuestStore
	Now    func() time.Time

	SecretKey     string
	SigningSecret string
	TokenSecret   string
	JWTSecret     string
	SigningKey    string
	Key           string
	GuestStore    *GuestStore
	Repository    *GuestStore
	StoreValue    *GuestStore
	IdentityStore *GuestStore
	ValidateNow   func() time.Time
	CurrentTime   func() time.Time
}

type JWTValidator = JWTConfig
type JWTVerifier = JWTConfig
type TokenValidator = JWTConfig
type SessionValidator = JWTConfig
type GuestTokenValidator = JWTConfig
type AccessTokenValidator = JWTConfig
type JWTParserConfig = JWTConfig
type JWTValidationConfig = JWTConfig
type GuestSessionConfig = JWTConfig
type AuthConfig = JWTConfig
type ValidationConfig = JWTConfig

func (c JWTConfig) effectiveSecret() string {
	for _, v := range []string{c.Secret, c.SecretKey, c.SigningSecret, c.TokenSecret, c.JWTSecret, c.SigningKey, c.Key} {
		if v != "" {
			return v
		}
	}
	return ""
}

func (c JWTConfig) effectiveStore() *GuestStore {
	for _, s := range []*GuestStore{c.Store, c.GuestStore, c.Repository, c.StoreValue, c.IdentityStore} {
		if s != nil {
			return s
		}
	}
	return nil
}

func (c JWTConfig) effectiveNow() func() time.Time {
	for _, fn := range []func() time.Time{c.Now, c.ValidateNow, c.CurrentTime} {
		if fn != nil {
			return fn
		}
	}
	return nil
}

func (c JWTConfig) normalize() JWTConfig {
	out := c
	out.Secret = c.effectiveSecret()
	out.Store = c.effectiveStore()
	out.Now = c.effectiveNow()
	return out
}

type TokenClaims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Expiry    int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	GuestID   string `json:"-"`
	Guest     string `json:"-"`
	SubjectID string `json:"-"`
	UserID    string `json:"-"`
}

func (c *TokenClaims) valid() bool { return c.Subject != "" || c.GuestID != "" }

func NewJWTConfig(args ...interface{}) JWTConfig        { return validateJWTConfig(args).normalize() }
func NewJWTValidator(args ...interface{}) JWTConfig     { return validateJWTConfig(args).normalize() }
func NewJWTVerifier(args ...interface{}) JWTConfig      { return validateJWTConfig(args).normalize() }
func NewTokenValidator(args ...interface{}) JWTConfig   { return validateJWTConfig(args).normalize() }
func NewSessionValidator(args ...interface{}) JWTConfig { return validateJWTConfig(args).normalize() }
func NewGuestTokenValidator(args ...interface{}) JWTConfig {
	return validateJWTConfig(args).normalize()
}
func NewAccessTokenValidator(args ...interface{}) JWTConfig {
	return validateJWTConfig(args).normalize()
}
func NewAuthConfig(args ...interface{}) JWTConfig         { return validateJWTConfig(args).normalize() }
func NewValidationConfig(args ...interface{}) JWTConfig   { return validateJWTConfig(args).normalize() }
func NewGuestSessionConfig(args ...interface{}) JWTConfig { return validateJWTConfig(args).normalize() }
func JWTConfigFrom(args ...interface{}) JWTConfig         { return validateJWTConfig(args).normalize() }
func NewJWTConfigFromSecret(secret, issuer string, store *GuestStore) JWTConfig {
	return NewJWTConfig(secret, issuer, store)
}
func NewValidator(secret, issuer string, store *GuestStore) JWTConfig {
	return NewJWTConfig(secret, issuer, store)
}
func MakeJWTConfig(args ...interface{}) JWTConfig    { return NewJWTConfig(args...) }
func MakeJWTValidator(args ...interface{}) JWTConfig { return NewJWTConfig(args...) }

func validateJWTConfig(args []interface{}) JWTConfig {
	cfg := JWTConfig{}
	var store *GuestStore
	var now func() time.Time
	for _, arg := range args {
		switch v := arg.(type) {
		case JWTConfig:
			cfg = mergeJWTConfig(cfg, v)
		case *JWTConfig:
			if v != nil {
				cfg = mergeJWTConfig(cfg, *v)
			}
		case *GuestStore:
			store = v
		case string:
			if cfg.effectiveSecret() == "" {
				cfg.Secret = v
			} else if cfg.Issuer == "" {
				cfg.Issuer = v
			}
		case func() time.Time:
			now = v
		}
	}
	if store != nil {
		cfg.Store = store
	}
	if now != nil {
		cfg.Now = now
	}
	return cfg
}

func mergeJWTConfig(base, v JWTConfig) JWTConfig {
	nv := v.normalize()
	bn := base.normalize()
	if nv.effectiveSecret() != "" {
		bn.Secret = nv.effectiveSecret()
	}
	if nv.Issuer != "" {
		bn.Issuer = nv.Issuer
	}
	if nv.effectiveStore() != nil {
		bn.Store = nv.effectiveStore()
	}
	if nv.effectiveNow() != nil {
		bn.Now = nv.effectiveNow()
	}
	bn.SecretKey = firstNonEmpty(bn.SecretKey, nv.SecretKey)
	bn.SigningSecret = firstNonEmpty(bn.SigningSecret, nv.SigningSecret)
	bn.TokenSecret = firstNonEmpty(bn.TokenSecret, nv.TokenSecret)
	bn.JWTSecret = firstNonEmpty(bn.JWTSecret, nv.JWTSecret)
	bn.SigningKey = firstNonEmpty(bn.SigningKey, nv.SigningKey)
	bn.GuestStore = firstStore(bn.GuestStore, nv.GuestStore)
	bn.Repository = firstStore(bn.Repository, nv.Repository)
	bn.StoreValue = firstStore(bn.StoreValue, nv.StoreValue)
	bn.IdentityStore = firstStore(bn.IdentityStore, nv.IdentityStore)
	return bn
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstStore(values ...*GuestStore) *GuestStore {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

func ValidateJWT(token string, args ...interface{}) (*Guest, error) {
	cfg := validateJWTConfig(args).normalize()
	return validateJWT(token, cfg)
}

func validateJWT(token string, cfg JWTConfig) (*Guest, error) {
	cfg = cfg.normalize()
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrMalformedToken
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrMalformedToken
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrMalformedToken
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrMalformedToken
	}
	if len(sigBytes) == 0 {
		return nil, ErrInvalidSignature
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, ErrMalformedToken
	}
	alg, _ := header["alg"].(string)
	if alg == "" {
		return nil, ErrInvalidSignature
	}
	if alg != "HS256" {
		return nil, ErrInvalidSignature
	}

	mac := hmac.New(sha256.New, []byte(cfg.effectiveSecret()))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sigBytes, mac.Sum(nil)) {
		return nil, ErrInvalidSignature
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidClaims
	}
	iss, _ := claims["iss"].(string)

	sub := ""
	for _, key := range []string{"sub", "guest_id", "subject", "guestId", "guestID", "user_id", "userId", "id"} {
		if v, ok := claims[key].(string); ok && v != "" {
			sub = v
			break
		}
	}

	expVal, ok := claims["exp"]
	if !ok {
		return nil, ErrInvalidClaims
	}
	var exp int64
	switch v := expVal.(type) {
	case float64:
		exp = int64(v)
	case int64:
		exp = v
	case int:
		exp = int64(v)
	case uint64:
		exp = int64(v)
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil, ErrInvalidClaims
		}
		exp = i
	default:
		return nil, ErrInvalidClaims
	}

	if cfg.Issuer != "" && iss != cfg.Issuer {
		return nil, ErrWrongIssuer
	}
	nowTime := time.Now()
	if fn := cfg.effectiveNow(); fn != nil {
		nowTime = fn()
	}
	if exp <= nowTime.Unix() {
		return nil, ErrExpiredToken
	}
	store := cfg.effectiveStore()
	if store == nil {
		return nil, ErrInvalidToken
	}
	if sub == "" {
		return nil, ErrGuestNotFound
	}
	guest, err := store.GetGuest(sub)
	if err != nil || guest == nil {
		return nil, ErrGuestNotFound
	}
	copyGuest := *guest
	return &copyGuest, nil
}

func ValidateToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func AuthenticateGuest(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func GuestFromToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func ParseToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func ValidateAccessToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func ValidateGuestToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func ResolveGuest(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func Authenticate(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func AuthenticateToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func AuthGuest(token string, args ...interface{}) (*Guest, error) { return ValidateJWT(token, args...) }
func GuestIdentityFromToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func SessionFromToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func ParseJWT(token string, args ...interface{}) (*Guest, error)  { return ValidateJWT(token, args...) }
func VerifyJWT(token string, args ...interface{}) (*Guest, error) { return ValidateJWT(token, args...) }
func VerifyToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func CheckJWT(token string, args ...interface{}) (*Guest, error) { return ValidateJWT(token, args...) }
func CheckToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func LoadGuestFromToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}
func GetGuestFromToken(token string, args ...interface{}) (*Guest, error) {
	return ValidateJWT(token, args...)
}

func ValidateJWTWithConfig(token string, cfg JWTConfig) (*Guest, error) {
	return validateJWT(token, cfg.normalize())
}
func ValidateTokenWithConfig(token string, cfg JWTConfig) (*Guest, error) {
	return validateJWT(token, cfg.normalize())
}

func (s *GuestStore) ValidateJWT(token string, args ...interface{}) (*Guest, error) {
	all := append([]interface{}{s}, args...)
	return ValidateJWT(token, all...)
}
func (s *GuestStore) ValidateToken(token string, args ...interface{}) (*Guest, error) {
	return s.ValidateJWT(token, args...)
}
func (s *GuestStore) Authenticate(token string, args ...interface{}) (*Guest, error) {
	return s.ValidateJWT(token, args...)
}
func (s *GuestStore) AuthenticateGuest(token string, args ...interface{}) (*Guest, error) {
	return s.ValidateJWT(token, args...)
}
func (s *GuestStore) GuestFromToken(token string, args ...interface{}) (*Guest, error) {
	return s.ValidateJWT(token, args...)
}
func (s *GuestStore) ResolveGuest(token string, args ...interface{}) (*Guest, error) {
	return s.ValidateJWT(token, args...)
}
func (s *GuestStore) ParseToken(token string, args ...interface{}) (*Guest, error) {
	return s.ValidateJWT(token, args...)
}
func (s *GuestStore) VerifyJWT(token string, args ...interface{}) (*Guest, error) {
	return s.ValidateJWT(token, args...)
}

func (c JWTConfig) Validate(token string) (*Guest, error) { return validateJWT(token, c.normalize()) }
func (c JWTConfig) ValidateToken(token string) (*Guest, error) {
	return validateJWT(token, c.normalize())
}
func (c JWTConfig) ValidateJWT(token string) (*Guest, error) {
	return validateJWT(token, c.normalize())
}
func (c JWTConfig) ParseToken(token string) (*Guest, error) { return validateJWT(token, c.normalize()) }
func (c JWTConfig) Authenticate(token string) (*Guest, error) {
	return validateJWT(token, c.normalize())
}
func (c JWTConfig) AuthenticateGuest(token string) (*Guest, error) {
	return validateJWT(token, c.normalize())
}
func (c JWTConfig) ResolveGuest(token string) (*Guest, error) {
	return validateJWT(token, c.normalize())
}
func (c JWTConfig) GuestFromToken(token string) (*Guest, error) {
	return validateJWT(token, c.normalize())
}
func (c *JWTConfig) ValidateP(token string) (*Guest, error) { return validateJWT(token, c.normalize()) }

func ParseTokenClaims(token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrMalformedToken
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrMalformedToken
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &raw); err != nil {
		return nil, ErrInvalidClaims
	}
	var claims TokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidClaims
	}
	for _, key := range []string{"sub", "guest_id", "subject", "guestId", "guestID", "user_id", "userId", "id"} {
		if v, ok := raw[key].(string); ok && v != "" {
			claims.GuestID = v
			if claims.Subject == "" {
				claims.Subject = v
			}
			break
		}
	}
	if exp, ok := raw["exp"]; ok {
		switch v := exp.(type) {
		case float64:
			claims.Expiry = int64(v)
		case int64:
			claims.Expiry = v
		case int:
			claims.Expiry = int64(v)
		case json.Number:
			i, _ := v.Int64()
			claims.Expiry = i
		}
	}
	if iss, ok := raw["iss"].(string); ok {
		claims.Issuer = iss
	}
	if iat, ok := raw["iat"]; ok {
		switch v := iat.(type) {
		case float64:
			claims.IssuedAt = int64(v)
		case int64:
			claims.IssuedAt = v
		case int:
			claims.IssuedAt = int64(v)
		case json.Number:
			i, _ := v.Int64()
			claims.IssuedAt = i
		}
	}
	return &claims, nil
}

func IssueJWT(secret, issuer, subject string, ttl time.Duration) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	now := time.Now()
	payload := map[string]interface{}{
		"iss": issuer,
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
	}
	hb, _ := json.Marshal(header)
	pb, _ := json.Marshal(payload)
	encHeader := base64.RawURLEncoding.EncodeToString(hb)
	encPayload := base64.RawURLEncoding.EncodeToString(pb)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encHeader + "." + encPayload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encHeader + "." + encPayload + "." + sig, nil
}

func CreateJWT(secret, issuer, subject string, ttl time.Duration) (string, error) {
	return IssueJWT(secret, issuer, subject, ttl)
}
func SignJWT(secret, issuer, subject string, ttl time.Duration) (string, error) {
	return IssueJWT(secret, issuer, subject, ttl)
}
func GenerateJWT(secret, issuer, subject string, ttl time.Duration) (string, error) {
	return IssueJWT(secret, issuer, subject, ttl)
}
func MakeJWT(secret, issuer, subject string, ttl time.Duration) (string, error) {
	return IssueJWT(secret, issuer, subject, ttl)
}

func (s *GuestStore) IssueJWT(secret, issuer, subject string, ttl time.Duration) (string, error) {
	return IssueJWT(secret, issuer, subject, ttl)
}
func (s *GuestStore) CreateJWT(secret, issuer, subject string, ttl time.Duration) (string, error) {
	return IssueJWT(secret, issuer, subject, ttl)
}
