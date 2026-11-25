/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package store

import (
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// applyTransforms runs write-time field transforms declared via `otmc:"hash=..."`.
// Empty values and already-hashed values are left untouched, so PATCH requests
// that omit the field (or resend the stored hash) never double-hash.
func (r *Repository) applyTransforms(inst any) error {
	ev := reflect.ValueOf(inst).Elem()
	for _, f := range r.model.Fields {
		if f.Hash == "" {
			continue
		}
		fv := ev.FieldByName(f.GoName)
		if !fv.IsValid() || fv.Kind() != reflect.String || !fv.CanSet() {
			continue
		}
		plain := fv.String()
		if plain == "" || isHashedValue(f.Hash, plain) {
			continue
		}
		hashed, err := hashValue(f.Hash, plain)
		if err != nil {
			return fmt.Errorf("otmc: hash %s: %w", f.JSONName, err)
		}
		fv.SetString(hashed)
	}
	return nil
}

// hashValue hashes plaintext with the named algorithm.
func hashValue(algo, plain string) (string, error) {
	switch strings.ToLower(algo) {
	case "bcrypt":
		b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
		return string(b), err
	default:
		return "", fmt.Errorf("unsupported hash algorithm %q", algo)
	}
}

// isHashedValue reports whether value already looks like output of algo,
// so stored hashes are never re-hashed on update.
func isHashedValue(algo, value string) bool {
	switch strings.ToLower(algo) {
	case "bcrypt":
		return strings.HasPrefix(value, "$2")
	default:
		return false
	}
}
