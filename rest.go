/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package rest

import (
	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/mapper"
	"github.com/otmc-sw/rest/pipeline"
	"github.com/otmc-sw/rest/validator"
)

type Context = context.Context

type Handler[Req any, Entity any] = pipeline.Handler[Req, Entity]

type UpdateHandler[Req any, Entity any] = pipeline.UpdateHandler[Req, Entity]

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
