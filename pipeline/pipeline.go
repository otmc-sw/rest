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

type Handler[Req any, Entity any] func(ctx context.Context, req Req) (Entity, error)

type UpdateHandler[Req any, Entity any] func(ctx context.Context, id string, req Req) (Entity, error)

type ExecHandler[Req any] func(ctx context.Context, req Req) error

type ExecHandlerWithID[Req any] func(ctx context.Context, req Req, id int64) error

type Pipeline[Req any, Entity any, Res any] struct {
	ctx      context.Context
	id       string
	parsedID int64
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
	debugger.PipelineStep("Param", "key=%s value=%s", key, p.id)
	return p
}

func (p *Pipeline[Req, Entity, Res]) ID() string {
	return p.id
}

func (p *Pipeline[Req, Entity, Res]) IntID() *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.id == "" {
		return p
	}
	id, err := strconv.ParseInt(p.id, 10, 64)
	if err != nil {
		debugger.Pipeline("IntID parse error: %v", err)
		p.bindErr = err
		return p
	}
	p.parsedID = id
	debugger.Pipeline("IntID parsed: %d", id)
	return p
}

func (p *Pipeline[Req, Entity, Res]) IDInt() int64 {
	return p.parsedID
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

func (p *Pipeline[Req, Entity, Res]) Handle(handler Handler[Req, Entity]) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	debugger.PipelineStep("Handle", "executing handler")
	entity, err := handler(p.ctx, *p.bound)
	if err != nil {
		debugger.Pipeline("Handle error: %v", err)
		p.bindErr = err
		return p
	}
	debugger.Pipeline("Handle success")
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

func (p *Pipeline[Req, Entity, Res]) SetEntity(entity Entity) *Pipeline[Req, Entity, Res] {
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) SetEntityFn(fn func() Entity) *Pipeline[Req, Entity, Res] {
	p.entityFn = fn
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
	if err := handler(p.ctx, *p.bound); err != nil {
		p.bindErr = err
		return p
	}
	entity := mapper.Map[Entity](*p.bound)
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) ExecWithID(handler ExecHandlerWithID[Req]) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	if p.id == "" {
		p.id = request.Param(p.ctx, "id")
	}
	if p.parsedID == 0 && p.id != "" {
		id, err := strconv.ParseInt(p.id, 10, 64)
		if err != nil {
			p.bindErr = err
			return p
		}
		p.parsedID = id
	}
	if err := handler(p.ctx, *p.bound, p.parsedID); err != nil {
		p.bindErr = err
		return p
	}
	entity := mapper.Map[Entity](*p.bound)
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) ExecWithIDResult(handler func(ctx context.Context, req Req, id int64) (any, error)) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	if p.id == "" {
		p.id = request.Param(p.ctx, "id")
	}
	if p.parsedID == 0 && p.id != "" {
		id, err := strconv.ParseInt(p.id, 10, 64)
		if err != nil {
			p.bindErr = err
			return p
		}
		p.parsedID = id
	}
	result, err := handler(p.ctx, *p.bound, p.parsedID)
	if err != nil {
		p.bindErr = err
		return p
	}
	entity := mapper.Map[Entity](result)
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) ExecWithIDResultTyped(handler func(ctx context.Context, req Req, id int64) (Entity, error)) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	if p.id == "" {
		p.id = request.Param(p.ctx, "id")
	}
	if p.parsedID == 0 && p.id != "" {
		id, err := strconv.ParseInt(p.id, 10, 64)
		if err != nil {
			p.bindErr = err
			return p
		}
		p.parsedID = id
	}
	entity, err := handler(p.ctx, *p.bound, p.parsedID)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) ExecResult(handler func(ctx context.Context, req Req) (any, error)) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	result, err := handler(p.ctx, *p.bound)
	if err != nil {
		p.bindErr = err
		return p
	}
	entity := mapper.Map[Entity](result)
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) ExecResultTypedSlice(handler func(ctx context.Context, req Req) (Entity, error)) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	entity, err := handler(p.ctx, *p.bound)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) ExecResultTyped(handler func(ctx context.Context, req Req) (Entity, error)) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	entity, err := handler(p.ctx, *p.bound)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.entity = &entity
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
