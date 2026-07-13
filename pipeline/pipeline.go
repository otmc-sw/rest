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

type PatchHandler[Req any, Params any] func(ctx context.Context, req Req, params Params, id any) (any, error)

type Pipeline[Req any, Params any, Entity any, Res any] struct {
	ctx      context.Context
	id       any // string or int64
	bound    *Req
	entity   *Entity
	entityFn func() Entity
	bindErr  error
	status   int
	paramsFn func(Req) Params
	params   Params
}

func newPipeline[Req any, Params any, Entity any, Res any](ctx context.Context, status int) *Pipeline[Req, Params, Entity, Res] {
	return &Pipeline[Req, Params, Entity, Res]{ctx: ctx, status: status}
}

func Create[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Create[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 201)
}

func Get[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Get[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200)
}

func Update[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Update[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200)
}

func Delete[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Delete start")
	return newPipeline[Req, Params, Entity, Res](ctx, 204)
}

func Patch[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Patch[%T, %T, %T] start", *new(Req), *new(Params), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200)
}

func (p *Pipeline[Req, Params, Entity, Res]) Param(key string) *Pipeline[Req, Params, Entity, Res] {
	p.id = request.Param(p.ctx, key)
	debugger.PipelineStep("Param", "key=%s value=%v", key, p.id)
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Bind() *Pipeline[Req, Params, Entity, Res] {
	var req Req
	debugger.PipelineStep("Bind", "binding request")
	if err := request.Bind(p.ctx, &req); err != nil {
		debugger.Pipeline("Bind error: %v", err)
		debugger.Error(debugger.ComponentPipeline, "Bind error: %v", err)
		return &Pipeline[Req, Params, Entity, Res]{ctx: p.ctx, bindErr: err, status: p.status}
	}
	debugger.Pipeline("Bind success: %+v", req)
	return &Pipeline[Req, Params, Entity, Res]{ctx: p.ctx, bound: &req, status: p.status}
}

func (p *Pipeline[Req, Params, Entity, Res]) Validate(fn func(req Req) error) *Pipeline[Req, Params, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	debugger.PipelineStep("Validate", "validating request")
	if err := fn(*p.bound); err != nil {
		p.bindErr = err
	} else {
		debugger.Pipeline("Validate success")
	}
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Params(fn func(req Req) Params) *Pipeline[Req, Params, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	debugger.PipelineStep("Params", "building params from request")
	p.paramsFn = fn
	p.params = fn(*p.bound)
	debugger.Pipeline("Params success: %+v", p.params)
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Handle(handler Handler[Req, Entity]) *Pipeline[Req, Params, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	p.ensureID()
	debugger.PipelineStep("Handle", "executing handler")
	entity, err := handler(p.ctx, *p.bound, p.id)
	if err != nil {
		debugger.Pipeline("Handle error: %v", err)
		debugger.Error(debugger.ComponentPipeline, "Handle error: %v", err)
		p.bindErr = err
		return p
	}
	debugger.Pipeline("Handle success")
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Exec(handler PatchHandler[Req, Params]) *Pipeline[Req, Params, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	p.ensureID()
	if p.paramsFn == nil {
		p.params = mapper.Map[Params](*p.bound)
		debugger.PipelineStep("Exec", "auto-converted Req->Params: %+v", p.params)
	}
	debugger.PipelineStep("Exec", "executing handler")
	result, err := handler(p.ctx, *p.bound, p.params, p.id)
	if err != nil {
		debugger.Pipeline("Exec error: %v", err)
		debugger.Error(debugger.ComponentPipeline, "Exec error: %v", err)
		p.bindErr = err
		return p
	}
	debugger.Pipeline("Exec success")
	if result != nil {
		entity := mapper.Map[Entity](result)
		p.entity = &entity
	} else {
		var entity Entity
		if p.paramsFn != nil {
			entity = mapper.Map[Entity](p.params)
		} else {
			entity = mapper.Map[Entity](*p.bound)
		}
		p.entity = &entity
	}
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Respond() error {
	debugger.PipelineStep("Respond", "preparing response (status=%d)", p.status)

	if p.bindErr != nil {
		debugger.Error(debugger.ComponentPipeline, "📚 Reason : %v", p.bindErr)
		if appErr, ok := p.bindErr.(errors.Error); ok {
			return errors.New().Skip(2).
				Code(appErr.Details.Code).
				Summary("Request Failed").
				Detail(p.bindErr).
				Send(p.ctx)
		}
		return errors.New().Skip(2).BadRequest().Summary("Request Failed").Detail(p.bindErr).Send(p.ctx)
	}

	if p.entityFn != nil {
		entity := p.entityFn()
		p.entity = &entity
		debugger.Pipeline("Respond using entityFn")
	}

	if p.entity == nil {
		debugger.Pipeline("Respond: no result produced")
		debugger.Error(debugger.ComponentPipeline, "Respond: no result produced")
		return errors.New().Skip(2).InternalError().Summary("no result produced").Send(p.ctx)
	}

	res := mapper.Map[Res](*p.entity)
	debugger.Pipeline("Respond success")
	return response.New[Res](p.ctx, p.status).Data(res).Send()
}

func (p *Pipeline[Req, Params, Entity, Res]) autoParseID() {
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

func (p *Pipeline[Req, Params, Entity, Res]) ensureID() {
	if p.id == nil {
		p.id = request.Param(p.ctx, "id")
	}
	p.autoParseID()
}

func Validate() *validator.Validator {
	return validator.New()
}
