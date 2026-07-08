/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package response

import (
	"github.com/otmc-sw/rest/errors"
)

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

type Builder struct {
	statusCode int
	data       interface{}
	message    string
	errBuilder *errors.Builder
}

func OK() *Builder {
	return &Builder{
		statusCode: 200,
	}
}

func Created() *Builder {
	return &Builder{
		statusCode: 201,
	}
}

func Accepted() *Builder {
	return &Builder{
		statusCode: 202,
	}
}

func NoContent() *Builder {
	return &Builder{
		statusCode: 204,
	}
}

func Error() *errors.Builder {
	return errors.New()
}

func (b *Builder) Data(data interface{}) *Builder {
	b.data = data
	return b
}

func (b *Builder) Message(msg string) *Builder {
	b.message = msg
	return b
}

func (b *Builder) Build() SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    b.data,
		Message: b.message,
	}
}

func (b *Builder) StatusCode() int {
	return b.statusCode
}
