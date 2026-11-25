/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/

// Package auth provides shared JWT/OIDC primitives for OTMC applications:
//
//   - Server side (IdP): KeyManager loads or generates an RSA keypair,
//     signs access/id tokens and exposes a JWKS endpoint.
//   - Client side (resource servers): Provider discovers the authority's
//     JWKS URI, caches the keys and verifies Bearer tokens. Apps only need
//     the authority URL - RequireAuth wires it into Fiber as middleware.
package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims are the standard token claims issued by OTMC Authenticator.
// They embed jwt.RegisteredClaims (iss, sub, aud, exp, iat, nonce is custom).
type Claims struct {
	jwt.RegisteredClaims
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Name              string   `json:"name,omitempty"`
	Email             string   `json:"email,omitempty"`
	EmailVerified     bool     `json:"email_verified,omitempty"`
	Picture           string   `json:"picture,omitempty"`
	Roles             []string `json:"roles,omitempty"`
	Nonce             string   `json:"nonce,omitempty"`
}

// Config configures a remote Provider (resource-server side).
type Config struct {
	// Authority is the IdP base URL, e.g. "http://localhost:5007".
	// Discovery is lazy: GET {Authority}/.well-known/openid-configuration.
	Authority string

	// Audience, when set, requires the token's aud claim to contain it.
	Audience string

	// Issuer overrides the expected iss claim; defaults to Authority.
	Issuer string
}
