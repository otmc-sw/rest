/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package pipeline

import (
	"strconv"

	"github.com/otmc-sw/rest/context"
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
	ctx       context.Context
	id        string
	parsedID  int64
	bound     *Req
	entity    *Entity
	entityFn  func() Entity
	bindErr   error
	status    int
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

// IntID parses the id string into int64 and stores it.
// If parsing fails, it sets bindErr so subsequent steps are skipped.
func (p *Pipeline[Req, Entity, Res]) IntID() *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.id == "" {
		return p
	}
	id, err := strconv.ParseInt(p.id, 10, 64)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.parsedID = id
	return p
}

// IDInt returns the parsed int64 id.
func (p *Pipeline[Req, Entity, Res]) IDInt() int64 {
	return p.parsedID
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

// SetEntity sets the entity directly. Used with Exec/ExecWithID when the handler
// needs to return a specific entity (e.g. from a database query) instead of auto-mapping.
func (p *Pipeline[Req, Entity, Res]) SetEntity(entity Entity) *Pipeline[Req, Entity, Res] {
	p.entity = &entity
	return p
}

// SetEntityFn sets a lazy function that returns the entity at Respond time.
// Used when the entity depends on data fetched inside ExecWithID.
func (p *Pipeline[Req, Entity, Res]) SetEntityFn(fn func() Entity) *Pipeline[Req, Entity, Res] {
	p.entityFn = fn
	return p
}

// Exec runs a handler that only returns an error.
// On success, it auto-maps Req -> Entity using the mapper.
// On error, it stores the error for Respond to handle.
func (p *Pipeline[Req, Entity, Res]) Exec(handler ExecHandler[Req]) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	if err := handler(p.ctx, *p.bound); err != nil {
		p.bindErr = err
		return p
	}
	entity := mapper.Map[Entity](*p.bound)
	p.entity = &entity
	return p
}

// ExecWithID runs a handler that receives the parsed int64 id and only returns an error.
// On success, it auto-maps Req -> Entity using the mapper.
// On error, it stores the error for Respond to handle.
func (p *Pipeline[Req, Entity, Res]) ExecWithID(handler ExecHandlerWithID[Req]) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	if err := handler(p.ctx, *p.bound, p.parsedID); err != nil {
		p.bindErr = err
		return p
	}
	entity := mapper.Map[Entity](*p.bound)
	p.entity = &entity
	return p
}

// ExecWithIDResult runs a handler that returns (Entity, error) along with the parsed id.
// On error, stores the error for Respond to handle.
func (p *Pipeline[Req, Entity, Res]) ExecWithIDResult(handler func(ctx context.Context, req Req, id int64) (Entity, error)) *Pipeline[Req, Entity, Res] {
	if p.bindErr != nil || p.bound == nil {
		return p
	}
	entity, err := handler(p.ctx, *p.bound, p.parsedID)
	if err != nil {
		p.bindErr = err
		return p
	}
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Entity, Res]) Respond() error {
	if p.bindErr != nil {
		// If the error is an errors.Error, use its status code
		if appErr, ok := p.bindErr.(errors.Error); ok {
			return errors.New().
				Code(appErr.Details.Code).
				Summary("request failed").
				Detail(p.bindErr).
				Send(p.ctx)
		}
		return errors.New().BadRequest().Summary("request failed").Detail(p.bindErr).Send(p.ctx)
	}

	// Evaluate lazy entity function if set
	if p.entityFn != nil {
		entity := p.entityFn()
		p.entity = &entity
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