/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/

package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultKeyID = "otmc-key-1"
	keyBits      = 2048
)

// KeyManager owns the RSA keypair used to sign tokens (IdP side).
// Keys are persisted as PEM files so restarts keep the same kid.
type KeyManager struct {
	mu         sync.RWMutex
	privateKey *rsa.PrivateKey
	KeyID      string
	dir        string
}

// LoadOrCreateKeys loads the PEM private key from dir, generating a new RSA
// keypair on first run. The file is private.pem (0600); the public key is
// derived from it and served via JWKS.
func LoadOrCreateKeys(dir string) (*KeyManager, error) {
	privPath := filepath.Join(dir, "private.pem")

	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return generateKeys(dir, privPath)
	}
	priv, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, fmt.Errorf("auth: parse private key: %w", err)
	}
	return &KeyManager{privateKey: priv, KeyID: defaultKeyID, dir: dir}, nil
}

func generateKeys(dir, privPath string) (*KeyManager, error) {
	priv, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, fmt.Errorf("auth: generate RSA key: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("auth: create key dir: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		return nil, fmt.Errorf("auth: write private key: %w", err)
	}
	return &KeyManager{privateKey: priv, KeyID: defaultKeyID, dir: dir}, nil
}

// PublicKey returns the RSA public key (for JWKS or local verification).
func (k *KeyManager) PublicKey() *rsa.PublicKey {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return &k.privateKey.PublicKey
}

// Sign signs claims with RS256 and stamps the kid header.
func (k *KeyManager) Sign(claims *Claims) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = k.KeyID
	return token.SignedString(k.privateKey)
}

// Verify parses and validates a token signed by this key manager
// (signature + registered claims such as exp). Issuer/audience are NOT
// checked here; callers compare claims.Issuer/claims.Audience as needed.
func (k *KeyManager) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return k.PublicKey(), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("auth: invalid token")
	}
	return claims, nil
}

// ---------------------------------------------------------------------------
// JWKS
// ---------------------------------------------------------------------------

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS returns the JSON Web Key Set document for this key manager.
func (k *KeyManager) JWKS() map[string]any {
	pub := k.PublicKey()
	return map[string]any{
		"keys": []jwk{{
			Kty: "RSA",
			Kid: k.KeyID,
			Alg: "RS256",
			Use: "sig",
			N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		}},
	}
}

// JWKSHandler serves GET /.well-known/jwks.json.
func JWKSHandler(k *KeyManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(k.JWKS())
	}
}

// ---------------------------------------------------------------------------
// Claim helpers
// ---------------------------------------------------------------------------

// NewAccessClaims builds standard access-token claims for a user.
func NewAccessClaims(issuer, subject string, roles []string, audience string, ttl time.Duration) *Claims {
	now := time.Now()
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Roles: roles,
	}
}

