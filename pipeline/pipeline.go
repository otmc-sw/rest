/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package pipeline

import (
	"strconv"

	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/debugger"
	"github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/mapper"
	"github.com/otmc-sw/rest/request"
	"github.com/otmc-sw/rest/response"
	"github.com/otmc-sw/rest/validator"
)

type Handler[Req any, Entity any] func(ctx context.Context, req Req, id any) (Entity, error)

type ExecHandler[Req any] func(ctx context.Context, req Req, id any) (any, error)

type Pipeline[Req any, Entity any, Res any] struct {
	ctx      context.Context
	id       any // string or int64
	bound    *Req
	entity   *Entity
	entityFn func() Entity
	bindErr  error
	status   int
}

func newPipeline[Req any, Entity any, Res any](ctx context.Context, status int) *Pipeline[Req, Entity, Res] {
	return &Pipeline[Req, Entity, Res]{ctx: ctx, status: status}
}

func Create[Req any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Entity, Res] {
	debugger.Pipeline("Create[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Entity, Res](ctx, 201)
}

func Get[Req any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Entity, Res] {
	debugger.Pipeline("Get[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Entity, Res](ctx, 200)
}

func Update[Req any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Entity, Res] {
	debugger.Pipeline("Update[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Entity, Res](ctx, 200)
}

func Delete[Res any](ctx context.Context) *Pipeline[struct{}, struct{}, Res] {
	debugger.Pipeline("Delete start")
	return newPipeline[struct{}, struct{}, Res](ctx, 204)
}

func (p *Pipeline[Req, Entity, Res]) Param(key string) *Pipeline[Req, Entity, Res] {
	p.id = request.Param(p.ctx, key)
	debugger.PipelineStep("Param", "key=%s value=%v", key, p.id)
	return p
}

func (p *Pipeline[Req, Entity, Res]) Bind() *Pipeline[Req, Entity, Res] {
	var req Req
	debugger.PipelineStep("Bind", "binding request")
	if err := request.Bind(p.ctx, &req); err != nil {
		debugger.Pipeline("Bind error: %v", err)
		return &Pipeline[Req, Entity, Res]{ctx: p.ctx, bindErr: err, status: p.status}
	}
	debugger.Pipeline("Bind success: %+v", req)
	return &Pipeline[Req, Entity, Res]{ctx: p.ctx, bound: &req, status: p.status}
}

func (p *Pipeline[Req, Entity, Res]) Validate(fn func(req Req) error) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	debugger.PipelineStep("Validate", "validating request")
	if err := fn(*p.bound); err != nil {
		debugger.Pipeline("Validate error: %v", err)
		p.bindErr = err
	} else {
		debugger.Pipeline("Validate success")
	}
	return p
}

func (p *Pipeline[Req, Entity, Res]) autoParseID() {
	if p.bindErr != nil {
		return
	}
	s, ok := p.id.(string)
	if !ok || s == "" {
		return
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		p.id = n
		debugger.Pipeline("autoParseID: parsed %q -> %d", s, n)
	}
}

func (p *Pipeline[Req, Entity, Res]) ensureID() {
	if p.id == nil {
		p.id = request.Param(p.ctx, "id")
	}
	p.autoParseID()
}

func (p *Pipeline[Req, Entity, Res]) Handle(handler Handler[Req, Entity]) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	p.ensureID()
	debugger.PipelineStep("Handle", "executing handler")
	entity, err := handler(p.ctx, *p.bound, p.id)
	if err != nil {
		debugger.Pipeline("Handle error: %v", err)
		p.bindErr = err
		return p
	}
	debugger.Pipeline("Handle success")
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) Exec(handler ExecHandler[Req]) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	p.ensureID()
	debugger.PipelineStep("Exec", "executing handler")
	result, err := handler(p.ctx, *p.bound, p.id)
	if err != nil {
		debugger.Pipeline("Exec error: %v", err)
		p.bindErr = err
		return p
	}
	debugger.Pipeline("Exec success")
	if result != nil {
		entity := mapper.Map[Entity](result)
		p.entity = &entity
	} else {
		entity := mapper.Map[Entity](*p.bound)
		p.entity = &entity
	}
	return p
}

func (p *Pipeline[Req, Entity, Res]) Respond() error {
	debugger.PipelineStep("Respond", "preparing response (status=%d)", p.status)

	if p.bindErr != nil {
		debugger.Pipeline("Respond error: %v", p.bindErr)
		if appErr, ok := p.bindErr.(errors.Error); ok {
			return errors.New().Skip(2).
				Code(appErr.Details.Code).
				Summary("request failed").
				Detail(p.bindErr).
				Send(p.ctx)
		}
		return errors.New().Skip(2).BadRequest().Summary("request failed").Detail(p.bindErr).Send(p.ctx)
	}

	if p.entityFn != nil {
		entity := p.entityFn()
		p.entity = &entity
		debugger.Pipeline("Respond using entityFn")
	}

	if p.entity == nil {
		debugger.Pipeline("Respond: no result produced")
		return errors.New().Skip(2).InternalError().Summary("no result produced").Send(p.ctx)
	}

	res := mapper.Map[Res](*p.entity)
	debugger.Pipeline("Respond success")
	return response.New[Res](p.ctx, p.status).Data(res).Send()
}

func Validate() *validator.Validator {
	return validator.New()
}
