/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
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

func Param(ctx context.Context, key string) string {
	return ctx.Param(key)
}

func ParamInt64(ctx context.Context, key string) (int64, error) {
	value := ctx.Param(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s is required", key)
	}
	return strconv.ParseInt(value, 10, 64)
}

func ParamInt(ctx context.Context, key string) (int, error) {
	value := ctx.Param(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s is required", key)
	}
	return strconv.Atoi(value)
}

func Query(ctx context.Context, key string) string {
	return ctx.Query(key)
}

func QueryInt64(ctx context.Context, key string) (int64, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

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

func QueryInt(ctx context.Context, key string) (int, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

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

func QueryBool(ctx context.Context, key string) (bool, error) {
	value := ctx.Query(key)
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

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

func QueryAll(ctx context.Context, key string) []string {
	return ctx.QueryAll(key)
}

func Header(ctx context.Context, key string) string {
	return ctx.Header(key)
}

func HeaderInt(ctx context.Context, key string) (int, error) {
	value := ctx.Header(key)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func Cookie(ctx context.Context, key string) string {
	return ctx.Cookie(key)
}

func Bind(ctx context.Context, v interface{}) error {
	return ctx.Bind(v)
}

func BindWithValidation(ctx context.Context, v interface{}, validateFn func(*validator.Validator)) error {
	if err := ctx.Bind(v); err != nil {
		return err
	}
	vld := validator.New()
	validateFn(vld)
	return vld.Process()
}

func String(ctx context.Context) (string, error) {
	return ctx.String()
}

func Bytes(ctx context.Context) ([]byte, error) {
	return ctx.Bytes()
}

func JSON(ctx context.Context, v interface{}) error {
	data, err := ctx.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

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

func GetContentType(ctx context.Context) string {
	return ctx.Header("Content-Type")
}

func GetAccept(ctx context.Context) string {
	return ctx.Header("Accept")
}

func GetUserAgent(ctx context.Context) string {
	return ctx.Header("User-Agent")
}

func GetClientIP(ctx context.Context) string {
	if xff := ctx.Header("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	if xri := ctx.Header("X-Real-IP"); xri != "" {
		return xri
	}
	return ctx.Header("X-Client-IP")
}
