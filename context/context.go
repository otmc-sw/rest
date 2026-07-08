/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package context

import (
	"context"
	"io"
)

type Context interface {
	Context() context.Context

	Param(key string) string

	Query(key string) string

	QueryAll(key string) []string

	Header(key string) string

	Cookie(key string) string

	Body() io.Reader

	Bind(v interface{}) error

	JSON(status int, body interface{}) error

	Status(code int)

	SetHeader(key, value string)

	Method() string

	Path() string

	String() (string, error)

	Bytes() ([]byte, error)
}
