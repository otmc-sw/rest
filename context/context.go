/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
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

	SendFile(path string) error

	Download(path string, name string) error

	HTML(html string) error

	Text(text string) error

	Redirect(location string) error

	Stream(reader io.Reader) error

	FormFile(key string) (interface{}, error)
}
