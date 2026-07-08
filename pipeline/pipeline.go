/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package pipeline

import (
	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/request"
	"github.com/otmc-sw/rest/response"
	"github.com/otmc-sw/rest/validator"
)

type Handler[Req any, Res any] func(ctx context.Context, req Req) (Res, error)

type UpdateHandler[Req any, Res any] func(ctx context.Context, id string, req Req) (Res, error)

func Create[Req any, Res any](ctx context.Context) *Pipeline[Req, Res] {
	return &Pipeline[Req, Res]{ctx: ctx}
}

func Update[Req any, Res any](ctx context.Context) *Pipeline[Req, Res] {
	return &Pipeline[Req, Res]{ctx: ctx}
}

type Pipeline[Req any, Res any] struct {
	ctx     context.Context
	id      string
	bound   *Req
	result  *Res
	bindErr error
}

func (p *Pipeline[Req, Res]) Param(key string) *Pipeline[Req, Res] {
	p.id = request.Param(p.ctx, key)
	return p
}

func (p *Pipeline[Req, Res]) Bind() *Pipeline[Req, Res] {
	var req Req
	if err := request.Bind(p.ctx, &req); err != nil {
		p.bindErr = err
		return p
	}
	p.bound = &req
	return p
}

func (p *Pipeline[Req, Res]) Validate(fn func(req Req) error) *Pipeline[Req, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	if err := fn(*p.bound); err != nil {
		p.bindErr = err
	}
	return p
}

func (p *Pipeline[Req, Res]) Handle(handler Handler[Req, Res]) *Pipeline[Req, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	res, err := handler(p.ctx, *p.bound)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.result = &res
	return p
}

func (p *Pipeline[Req, Res]) HandleWithID(handler UpdateHandler[Req, Res]) *Pipeline[Req, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	res, err := handler(p.ctx, p.id, *p.bound)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.result = &res
	return p
}

func (p *Pipeline[Req, Res]) Respond() error {
	if p.bindErr != nil {
		return errors.New().BadRequest().Summary("request failed").Detail(p.bindErr).Send(p.ctx)
	}
	if p.result == nil {
		return errors.New().InternalError().Summary("no result produced").Send(p.ctx)
	}
	return response.OK[Res](p.ctx).Data(*p.result).Send()
}

func Validate() *validator.Validator {
	return validator.New()
}
