/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package pipeline

import (
	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/mapper"
	"github.com/otmc-sw/rest/request"
	"github.com/otmc-sw/rest/response"
	"github.com/otmc-sw/rest/validator"
)

type Handler[Req any, Entity any] func(ctx context.Context, req Req) (Entity, error)

type UpdateHandler[Req any, Entity any] func(ctx context.Context, id string, req Req) (Entity, error)

type Pipeline[Req any, Entity any, Res any] struct {
	ctx     context.Context
	id      string
	bound   *Req
	entity  *Entity
	bindErr error
	status  int
}

func newPipeline[Req any, Entity any, Res any](ctx context.Context, status int) *Pipeline[Req, Entity, Res] {
	return &Pipeline[Req, Entity, Res]{ctx: ctx, status: status}
}

func Create[Req any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Entity, Res] {
	return newPipeline[Req, Entity, Res](ctx, 201)
}

func Get[Req any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Entity, Res] {
	return newPipeline[Req, Entity, Res](ctx, 200)
}

func Update[Req any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Entity, Res] {
	return newPipeline[Req, Entity, Res](ctx, 200)
}

func Delete[Res any](ctx context.Context) *Pipeline[struct{}, struct{}, Res] {
	return newPipeline[struct{}, struct{}, Res](ctx, 204)
}

func (p *Pipeline[Req, Entity, Res]) Param(key string) *Pipeline[Req, Entity, Res] {
	p.id = request.Param(p.ctx, key)
	return p
}

func (p *Pipeline[Req, Entity, Res]) ID() string {
	return p.id
}

func (p *Pipeline[Req, Entity, Res]) Bind() *Pipeline[Req, Entity, Res] {
	var req Req
	if err := request.Bind(p.ctx, &req); err != nil {
		return &Pipeline[Req, Entity, Res]{ctx: p.ctx, bindErr: err, status: p.status}
	}
	return &Pipeline[Req, Entity, Res]{ctx: p.ctx, bound: &req, status: p.status}
}

func (p *Pipeline[Req, Entity, Res]) Validate(fn func(req Req) error) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	if err := fn(*p.bound); err != nil {
		p.bindErr = err
	}
	return p
}

func (p *Pipeline[Req, Entity, Res]) Handle(handler Handler[Req, Entity]) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	entity, err := handler(p.ctx, *p.bound)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) HandleWithID(handler UpdateHandler[Req, Entity]) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	entity, err := handler(p.ctx, p.id, *p.bound)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) Respond() error {
	if p.bindErr != nil {
		return errors.New().BadRequest().Summary("request failed").Detail(p.bindErr).Send(p.ctx)
	}
	if p.entity == nil {
		return errors.New().InternalError().Summary("no result produced").Send(p.ctx)
	}

	res := mapper.Map[Res](*p.entity)
	return response.New[Res](p.ctx, p.status).Data(res).Send()
}

func Validate() *validator.Validator {
	return validator.New()
}