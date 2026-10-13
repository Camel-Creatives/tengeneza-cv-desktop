// Package auth owns credential hashing, OTP proofing, and product JWTs.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// VerifyTOTP implements RFC 6238 with the standard 30-second SHA-1 profile.
// Keeping this here avoids coupling admin authentication to a third-party QR library.
func VerifyTOTP(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	key := []byte(secret)
	if decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret))); err == nil {
		key = decoded
	}
	for _, offset := range []int64{-1, 0, 1} {
		counter := uint64(now.Unix()/30 + offset)
		var msg [8]byte
		for i := 7; i >= 0; i-- {
			msg[i] = byte(counter)
			counter >>= 8
		}
		h := hmac.New(sha1.New, key)
		_, _ = h.Write(msg[:])
		sum := h.Sum(nil)
		idx := sum[len(sum)-1] & 0x0f
		value := (uint32(sum[idx])&0x7f)<<24 | uint32(sum[idx+1])<<16 | uint32(sum[idx+2])<<8 | uint32(sum[idx+3])
		if fmt.Sprintf("%06d", value%1000000) == code {
			return true
		}
	}
	return false
}

type Manager struct {
	signingKey []byte
	issuer     string
}
type Claims struct {
	UserID      string `json:"user_id,omitempty"`
	Role        string `json:"role,omitempty"`
	Name        string `json:"name,omitempty"`
	FullName    string `json:"full_name,omitempty"`
	GivenName   string `json:"given_name,omitempty"`
	FamilyName  string `json:"family_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone_number,omitempty"`
	PhoneAlt    string `json:"phone,omitempty"`
	PhoneNumber string `json:"-"`
	jwt.RegisteredClaims
}
type OTP struct {
	ID, UserID, Destination, Purpose, CodeHash string
	ExpiresAt                                  time.Time
	Attempts                                   int
}

// Token lifetimes are centralized so handlers and tests cannot drift apart.
// Refresh tokens are opaque credentials and must be persisted server-side with
// this lifetime when the refresh endpoint is enabled.
const (
	AccessTokenTTL  = 7 * 24 * time.Hour
	RefreshTokenTTL = 30 * 24 * time.Hour
)

func New(signingKey, issuer string) *Manager {
	return &Manager{signingKey: []byte(signingKey), issuer: issuer}
}

func (m *Manager) IssueTokens(userID, role string) (string, string, error) {
	now := time.Now().UTC()
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{UserID: userID, Role: role, RegisteredClaims: jwt.RegisteredClaims{Issuer: m.issuer, Subject: userID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL))}})
	accessToken, err := access.SignedString(m.signingKey)
	if err != nil {
		return "", "", err
	}
	refreshRaw, err := randomToken(32)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshRaw, nil
}

func (m *Manager) ParseAccessToken(token string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.signingKey, nil
	})
	if err != nil {
		return Claims{}, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Claims{}, errors.New("invalid access token")
	}
	return *claims, nil
}

func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("password must contain at least 8 characters")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return fmt.Sprintf("argon2id$v=19$m=65536,t=3,p=2$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}
func VerifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" {
		return false
	}
	params := map[string]uint32{}
	for _, pair := range strings.Split(parts[2], ",") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return false
		}
		v, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return false
		}
		params[kv[0]] = uint32(v)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, params["t"], params["m"], uint8(params["p"]), uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}
func NewOTP(userID, destination, purpose string) (OTP, string, error) {
	code, err := numericCode()
	if err != nil {
		return OTP{}, "", err
	}
	return OTP{ID: uuid.NewString(), UserID: userID, Destination: destination, Purpose: purpose, CodeHash: otpHash(code), ExpiresAt: time.Now().UTC().Add(5 * time.Minute)}, code, nil
}
func VerifyOTP(otp OTP, code string) error {
	if time.Now().UTC().After(otp.ExpiresAt) {
		return errors.New("verification code expired")
	}
	if otp.Attempts >= 5 {
		return errors.New("verification is temporarily locked")
	}
	if subtle.ConstantTimeCompare([]byte(otp.CodeHash), []byte(otpHash(code))) != 1 {
		return errors.New("incorrect verification code")
	}
	return nil
}
func HashOpaqueToken(token string) string { return otpHash(token) }
func otpHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return base64.RawStdEncoding.EncodeToString(sum[:])
}
func randomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
func numericCode() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	value := (int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])) % 1000000
	return fmt.Sprintf("%06d", value), nil
}
