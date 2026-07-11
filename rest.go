/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package rest

import (
	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/debugger"
	"github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/mapper"
	"github.com/otmc-sw/rest/pipeline"
	"github.com/otmc-sw/rest/response"
	"github.com/otmc-sw/rest/validator"
)

type Context = context.Context

type Handler[Req any, Entity any] = pipeline.Handler[Req, Entity]

type ExecHandler[Req any] = pipeline.ExecHandler[Req]

type Pipeline[Req any, Entity any, Res any] = pipeline.Pipeline[Req, Entity, Res]

func Create[Req any, Entity any, Res any](ctx Context) *Pipeline[Req, Entity, Res] {
	return pipeline.Create[Req, Entity, Res](ctx)
}

func Get[Req any, Entity any, Res any](ctx Context) *Pipeline[Req, Entity, Res] {
	return pipeline.Get[Req, Entity, Res](ctx)
}

func Update[Req any, Entity any, Res any](ctx Context) *Pipeline[Req, Entity, Res] {
	return pipeline.Update[Req, Entity, Res](ctx)
}

func Delete[Res any](ctx Context) *Pipeline[struct{}, struct{}, Res] {
	return pipeline.Delete[Res](ctx)
}

func Register[Src any, Dst any](fn func(Src) Dst) {
	mapper.Register(fn)
}

func Validate() *validator.Validator {
	return validator.New()
}

func NewError() *errors.Builder {
	return errors.New()
}

func OK(ctx Context) *response.Builder[any] {
	return response.OK[any](ctx)
}

func BadRequest(summary string, err error) error {
	return errors.New().
		BadRequest().
		Skip(1).
		Summary(summary).
		Detail(err.Error()).
		Build()
}

func Unauthorized(summary string, err error) error {
	return errors.New().
		Unauthorized().
		Skip(1).
		Summary(summary).
		Detail(err.Error()).
		Build()
}

func Forbidden(summary string, err error) error {
	return errors.New().
		Forbidden().
		Skip(1).
		Summary(summary).
		Detail(err.Error()).
		Build()
}

func NotFound(summary string, err error) error {
	return errors.New().
		NotFound().
		Skip(1).
		Summary(summary).
		Detail(err.Error()).
		Build()
}

func Conflict(summary string, err error) error {
	return errors.New().
		Conflict().
		Skip(1).
		Summary(summary).
		Detail(err.Error()).
		Build()
}

func UnprocessableEntity(summary string, err error) error {
	return errors.New().
		UnprocessableEntity().
		Skip(1).
		Summary(summary).
		Detail(err.Error()).
		Build()
}

func InternalError(summary string, err error) error {
	return errors.New().
		InternalError().
		Skip(1).
		Summary(summary).
		Detail(err.Error()).
		Build()
}

func ServiceUnavailable(summary string, err error) error {
	return errors.New().
		ServiceUnavailable().
		Skip(1).
		Summary(summary).
		Detail(err.Error()).
		Build()
}

func Debug(enable ...bool) {
	if len(enable) > 0 && enable[0] {
		debugger.Enable()
	} else {
		debugger.Disable()
	}
}

func DebugComponent(component string) {
	debugger.EnableComponent(component)
}

func DebugWithEnv() {
	debugger.WithEnv()
}
