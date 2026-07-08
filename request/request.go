/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package request

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/validator"
)

// Param returns a path parameter by key
func Param(ctx context.Context, key string) string {
	return ctx.Param(key)
}

// ParamInt64 returns a path parameter as int64
func ParamInt64(ctx context.Context, key string) (int64, error) {
	value := ctx.Param(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s is required", key)
	}
	return strconv.ParseInt(value, 10, 64)
}

// ParamInt returns a path parameter as int
func ParamInt(ctx context.Context, key string) (int, error) {
	value := ctx.Param(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s is required", key)
	}
	return strconv.Atoi(value)
}

// Query returns a query parameter by key
func Query(ctx context.Context, key string) string {
	return ctx.Query(key)
}

// QueryInt64 returns a query parameter as int64
func QueryInt64(ctx context.Context, key string) (int64, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

// QueryInt64OrDefault returns a query parameter as int64 with default
func QueryInt64OrDefault(ctx context.Context, key string, defaultValue int64) int64 {
	value := ctx.Query(key)
	if value == "" {
		return defaultValue
	}
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return result
}

// QueryInt returns a query parameter as int
func QueryInt(ctx context.Context, key string) (int, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// QueryIntOrDefault returns a query parameter as int with default
func QueryIntOrDefault(ctx context.Context, key string, defaultValue int) int {
	value := ctx.Query(key)
	if value == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}

// QueryBool returns a query parameter as bool
func QueryBool(ctx context.Context, key string) (bool, error) {
	value := ctx.Query(key)
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

// QueryBoolOrDefault returns a query parameter as bool with default
func QueryBoolOrDefault(ctx context.Context, key string, defaultValue bool) bool {
	value := ctx.Query(key)
	if value == "" {
		return defaultValue
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return result
}

// QueryAll returns all query parameters by key
func QueryAll(ctx context.Context, key string) []string {
	return ctx.QueryAll(key)
}

// Header returns a header value by key
func Header(ctx context.Context, key string) string {
	return ctx.Header(key)
}

// HeaderInt returns a header value as int
func HeaderInt(ctx context.Context, key string) (int, error) {
	value := ctx.Header(key)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// Cookie returns a cookie value by key
func Cookie(ctx context.Context, key string) string {
	return ctx.Cookie(key)
}

// Bind binds the request body to a struct
func Bind(ctx context.Context, v interface{}) error {
	return ctx.Bind(v)
}

// BindWithValidation binds the request body to a struct and validates it
func BindWithValidation(ctx context.Context, v interface{}, validateFn func(*validator.Validator)) error {
	if err := ctx.Bind(v); err != nil {
		return err
	}
	vld := validator.New()
	validateFn(vld)
	return vld.Validate()
}

// String returns the request body as string
func String(ctx context.Context) (string, error) {
	return ctx.String()
}

// Bytes returns the request body as bytes
func Bytes(ctx context.Context) ([]byte, error) {
	return ctx.Bytes()
}

// JSON unmarshals the request body to a struct
func JSON(ctx context.Context, v interface{}) error {
	data, err := ctx.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// GetBearerToken extracts Bearer token from Authorization header
func GetBearerToken(ctx context.Context) (string, error) {
	auth := ctx.Header("Authorization")
	if auth == "" {
		return "", fmt.Errorf("authorization header is missing")
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return parts[1], nil
}

// GetContentType returns the Content-Type header
func GetContentType(ctx context.Context) string {
	return ctx.Header("Content-Type")
}

// GetAccept returns the Accept header
func GetAccept(ctx context.Context) string {
	return ctx.Header("Accept")
}

// GetUserAgent returns the User-Agent header
func GetUserAgent(ctx context.Context) string {
	return ctx.Header("User-Agent")
}

// GetClientIP returns the client IP address
func GetClientIP(ctx context.Context) string {
	// Try X-Forwarded-For header first
	if xff := ctx.Header("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	// Try X-Real-IP header
	if xri := ctx.Header("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to remote address
	return ctx.Header("X-Client-IP")
}
