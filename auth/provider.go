/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/

package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Provider verifies Bearer tokens issued by a remote OTMC Authenticator.
// It lazily discovers {Authority}/.well-known/openid-configuration, fetches
// the JWKS and caches RSA public keys. Unknown kid triggers one refetch,
// so key rotation is picked up without restart.
type Provider struct {
	cfg Config

	mu        sync.Mutex
	issuer    string
	jwksURI   string
	keys      map[string]*rsa.PublicKey
	disfed    bool
	lastFetch time.Time
	client    *http.Client
}

// NewProvider creates a provider. No network I/O happens until the first
// Verify call, so setup never fails on a down authority.
func NewProvider(cfg Config) *Provider {
	issuer := cfg.Issuer
	if issuer == "" {
		issuer = strings.TrimRight(cfg.Authority, "/")
	}
	return &Provider{
		cfg:    cfg,
		issuer: issuer,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Verify validates a Bearer token: signature (RS256 via JWKS), expiry and,
// when configured, issuer and audience.
func (p *Provider) Verify(ctx context.Context, tokenString string) (*Claims, error) {
	// Peek at the header to find kid before verifying.
	var head struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("auth: malformed token")
	}
	hdrJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("auth: malformed token header")
	}
	if err := json.Unmarshal(hdrJSON, &head); err != nil {
		return nil, errors.New("auth: malformed token header")
	}
	if head.Alg != "RS256" {
		return nil, fmt.Errorf("auth: unexpected alg %q", head.Alg)
	}

	key, err := p.keyFor(ctx, head.Kid, false)
	if err != nil {
		return nil, err
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("auth: invalid token")
	}
	if p.issuer != "" && claims.Issuer != p.issuer {
		return nil, fmt.Errorf("auth: unexpected issuer %q", claims.Issuer)
	}
	if p.cfg.Audience != "" {
		ok := false
		for _, a := range claims.Audience {
			if a == p.cfg.Audience {
				ok = true
				break
			}
		}
		if !ok {
			return nil, errors.New("auth: audience mismatch")
		}
	}
	return claims, nil
}

// keyFor returns the public key for kid, refetching JWKS once on miss.
func (p *Provider) keyFor(ctx context.Context, kid string, refetched bool) (*rsa.PublicKey, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.keys == nil {
		if err := p.refreshLocked(ctx); err != nil {
			return nil, err
		}
	}
	if k, ok := p.keys[kid]; ok {
		return k, nil
	}
	if !refetched {
		// Unknown kid: keys may have rotated - refetch once.
		p.keys = nil
		if err := p.refreshLocked(ctx); err != nil {
			return nil, err
		}
		if k, ok := p.keys[kid]; ok {
			return k, nil
		}
	}
	return nil, fmt.Errorf("auth: unknown key id %q", kid)
}

// refreshLocked performs discovery (once) + JWKS fetch. Caller holds p.mu.
func (p *Provider) refreshLocked(ctx context.Context) error {
	if !p.disfed {
		if err := p.discoverLocked(ctx); err != nil {
			return err
		}
		p.disfed = true
	}
	keys, err := p.fetchJWKS(ctx, p.jwksURI)
	if err != nil {
		return err
	}
	p.keys = keys
	p.lastFetch = time.Now()
	return nil
}

func (p *Provider) discoverLocked(ctx context.Context) error {
	url := strings.TrimRight(p.cfg.Authority, "/") + "/.well-known/openid-configuration"
	var doc struct {
		Issuer  string `json:"issuer"`
		JWKSURI string `json:"jwks_uri"`
	}
	if err := p.getJSON(ctx, url, &doc); err != nil {
		return fmt.Errorf("auth: discovery %s: %w", url, err)
	}
	if doc.JWKSURI == "" {
		return errors.New("auth: discovery document has no jwks_uri")
	}
	p.jwksURI = doc.JWKSURI
	return nil
}

// remoteJWK mirrors the public parts of a JWK.
type remoteJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (p *Provider) fetchJWKS(ctx context.Context, uri string) (map[string]*rsa.PublicKey, error) {
	var doc struct {
		Keys []remoteJWK `json:"keys"`
	}
	if err := p.getJSON(ctx, uri, &doc); err != nil {
		return nil, fmt.Errorf("auth: fetch jwks %s: %w", uri, err)
	}
	keys := map[string]*rsa.PublicKey{}
	for _, j := range doc.Keys {
		if j.Kty != "RSA" || j.N == "" || j.E == "" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 | int(b)
		}
		keys[j.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: e,
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("auth: jwks contains no usable RSA keys")
	}
	return keys, nil
}

func (p *Provider) getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
