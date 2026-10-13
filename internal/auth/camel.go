package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CamelVerifier struct {
	issuer, audience, jwksURL string
	client                    *http.Client
	mu                        sync.RWMutex
	keys                      map[string]*rsa.PublicKey
	expires                   time.Time
}
type jwksDocument struct {
	Keys []struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func NewCamelVerifier(issuer, audience, jwksURL string) *CamelVerifier {
	return &CamelVerifier{issuer: issuer, audience: audience, jwksURL: jwksURL, client: &http.Client{Timeout: 5 * time.Second}, keys: map[string]*rsa.PublicKey{}}
}

func (v *CamelVerifier) Verify(raw string) (Claims, error) {
	if strings.TrimSpace(raw) == "" {
		return Claims{}, errors.New("missing bearer token")
	}
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("unexpected signing method")
		}
		kid, _ := t.Header["kid"].(string)
		return v.key(kid)
	}, jwt.WithIssuer(v.issuer), jwt.WithAudience(v.audience))
	if err != nil || !token.Valid {
		return Claims{}, errors.New("invalid CamelAccounts access token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return Claims{}, errors.New("invalid token claims")
	}
	return *claims, nil
}

func (v *CamelVerifier) key(kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key, fresh := v.keys[kid], time.Now().Before(v.expires)
	v.mu.RUnlock()
	if fresh && key != nil {
		return key, nil
	}
	response, err := v.client.Get(v.jwksURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("CamelAccounts JWKS unavailable")
	}
	var doc jwksDocument
	if err := json.NewDecoder(response.Body).Decode(&doc); err != nil {
		return nil, err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, item := range doc.Keys {
		if item.Kty != "RSA" || item.Kid == "" {
			continue
		}
		n, e1 := base64.RawURLEncoding.DecodeString(item.N)
		e2, e := base64.RawURLEncoding.DecodeString(item.E)
		if e1 != nil || e != nil || len(e2) == 0 {
			continue
		}
		exponent := 0
		for _, b := range e2 {
			exponent = exponent<<8 | int(b)
		}
		keys[item.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exponent}
	}
	v.mu.Lock()
	v.keys, v.expires = keys, time.Now().Add(10*time.Minute)
	key = v.keys[kid]
	v.mu.Unlock()
	if key == nil {
		return nil, errors.New("unknown CamelAccounts signing key")
	}
	return key, nil
}
