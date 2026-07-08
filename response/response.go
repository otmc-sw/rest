/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package response

import (
	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/errors"
)

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

type Builder[T any] struct {
	ctx       context.Context
	statusCode int
	data       interface{}
	message    string
	errBuilder *errors.Builder
}

func OK[T any](ctx context.Context) *Builder[T] {
	return &Builder[T]{
		ctx:        ctx,
		statusCode: 200,
	}
}

func Created[T any](ctx context.Context) *Builder[T] {
	return &Builder[T]{
		ctx:        ctx,
		statusCode: 201,
	}
}

func Accepted[T any](ctx context.Context) *Builder[T] {
	return &Builder[T]{
		ctx:        ctx,
		statusCode: 202,
	}
}

func NoContent[T any](ctx context.Context) *Builder[T] {
	return &Builder[T]{
		ctx:        ctx,
		statusCode: 204,
	}
}

func Error() *errors.Builder {
	return errors.New()
}

func (b *Builder[T]) Data(data interface{}) *Builder[T] {
	b.data = data
	return b
}

func (b *Builder[T]) Message(msg string) *Builder[T] {
	b.message = msg
	return b
}

func (b *Builder[T]) Build() SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    b.data,
		Message: b.message,
	}
}

func (b *Builder[T]) StatusCode() int {
	return b.statusCode
}

func (b *Builder[T]) Send() error {
	if b.ctx == nil {
		return errors.New().InternalError().Summary("response context is nil").Build().Err()
	}
	return b.ctx.JSON(b.statusCode, b.Build())
}