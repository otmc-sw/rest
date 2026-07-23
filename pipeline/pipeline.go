/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package pipeline

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/otmc-sw/rest/config"
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
	ctx          context.Context
	id           any
	bound        *Req
	entity       *Entity
	entityFn     func() Entity
	err          error
	status       int
	paramsFn     func(Req) Params
	params       Params
	customFields map[string]any
	operation    string
}

func newPipeline[Req any, Params any, Entity any, Res any](ctx context.Context, status int) *Pipeline[Req, Params, Entity, Res] {
	return &Pipeline[Req, Params, Entity, Res]{ctx: ctx, status: status}
}

func Create[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Create[%T, %T]", *new(Req), *new(Entity))
	p := newPipeline[Req, Params, Entity, Res](ctx, 201)
	p.operation = "Post"
	return p
}

func Get[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Get[%T, %T]", *new(Req), *new(Entity))
	p := newPipeline[Req, Params, Entity, Res](ctx, 200)
	p.operation = "Get"
	return p
}

func Update[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Update[%T, %T]", *new(Req), *new(Entity))
	p := newPipeline[Req, Params, Entity, Res](ctx, 200)
	p.operation = "Update"
	return p
}

func Patch[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Patch[%T, %T, %T]", *new(Req), *new(Params), *new(Entity))
	p := newPipeline[Req, Params, Entity, Res](ctx, 200)
	p.operation = "Patch"
	return p
}

func Delete[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Delete")
	p := newPipeline[Req, Params, Entity, Res](ctx, 204)
	p.operation = "Delete"
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Param(key string) *Pipeline[Req, Params, Entity, Res] {
	if p.err != nil {
		return p
	}
	p.id = request.Param(p.ctx, key)
	debugger.PipelineStep("Param", "key=%s value=%v", key, p.id)
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Bind() *Pipeline[Req, Params, Entity, Res] {
	if p.err != nil {
		return p
	}
	var req Req
	debugger.PipelineStep("Bind", "binding request")
	if err := request.Bind(p.ctx, &req); err != nil {
		debugger.Error(debugger.ComponentPipeline, "Bind: %v", err)
		p.err = err
		return p
	}
	debugger.Pipeline("Bind success: %+v", req)
	p.bound = &req
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Validate(fn func(Req) error) *Pipeline[Req, Params, Entity, Res] {
	if p.err != nil || p.bound == nil {
		return p
	}
	debugger.PipelineStep("Validate", "validating")
	if err := fn(*p.bound); err != nil {
		p.err = err
		return p
	}
	debugger.Pipeline("Validate success")
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Params(fn func(Req) Params) *Pipeline[Req, Params, Entity, Res] {
	if p.err != nil || p.bound == nil {
		return p
	}
	debugger.PipelineStep("Params", "building from request")
	p.paramsFn = fn
	p.params = fn(*p.bound)
	debugger.Pipeline("Params: %+v", p.params)
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) SetFields(fields map[string]any) *Pipeline[Req, Params, Entity, Res] {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("SetFields", "fields=%+v", fields)
	p.customFields = fields
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) SetField(key string, value any) *Pipeline[Req, Params, Entity, Res] {
	if p.err != nil {
		return p
	}
	if p.customFields == nil {
		p.customFields = make(map[string]any)
	}
	debugger.PipelineStep("SetField", "key=%s value=%v", key, value)
	p.customFields[key] = value
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Handle(h Handler[Req, Entity]) *Pipeline[Req, Params, Entity, Res] {
	if p.err != nil || p.bound == nil {
		return p
	}
	p.ensureID()
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Handle", "executing")
	entity, err := h(p.ctx, *p.bound, p.id)
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "Handle: %v", err)
		p.err = err
		return p
	}
	debugger.Pipeline("Handle success")
	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Exec(h PatchHandler[Req, Params]) *Pipeline[Req, Params, Entity, Res] {
	if p.err != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	p.ensureID()
	if p.err != nil {
		return p
	}

	if p.paramsFn == nil {
		p.params = mapper.Map[Params](*p.bound)
		debugger.PipelineStep("Exec", "auto-mapped Req→Params: %+v", p.params)
	}

	mapper.SetField(&p.params, "ID", p.id)

	debugger.PipelineStep("Exec", "executing")
	result, err := h(p.ctx, *p.bound, p.params, p.id)
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "Exec: %v", err)
		p.err = err
		return p
	}
	debugger.Pipeline("Exec success")

	switch {
	case result != nil:
		e := mapper.Map[Entity](result)
		p.entity = &e
	case p.paramsFn != nil:
		e := mapper.Map[Entity](p.params)
		p.entity = &e
	default:
		e := mapper.Map[Entity](*p.bound)
		p.entity = &e
	}
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Respond() error {
	debugger.PipelineStep("Respond", "status=%d", p.status)

	if p.err != nil {
		return p.respondError()
	}
	if p.entityFn != nil {
		e := p.entityFn()
		p.entity = &e
	}
	if p.entity == nil {
		debugger.Error(debugger.ComponentPipeline, "Respond: no entity produced")
		return errors.New().Skip(2).InternalError().Summary("no result produced").Send(p.ctx)
	}

	res := mapper.Map[Res](*p.entity)

	applyFields := func(item any, source string) {
		globalCfg := config.GetGlobalConfig()
		if globalCfg == nil {
			return
		}

		preConfig := globalCfg.Pre()
		if preConfig != nil {
			fields := preConfig.GetFields()
			for key, value := range fields {
				setFieldAny(item, key, value)
				debugger.Pipeline("GlobalConfig[pre]: %s = %v", key, value)
			}
			fieldFuncs := preConfig.GetFieldFuncs()
			for key, fn := range fieldFuncs {
				value := fn(item)
				setFieldAny(item, key, value)
				debugger.Pipeline("GlobalConfigFunc[pre]: %s = %v", key, value)
			}
		}

		postConfig := globalCfg.Post()
		if postConfig != nil {
			fields := postConfig.GetFields()
			for key, value := range fields {
				setFieldAny(item, key, value)
				debugger.Pipeline("GlobalConfig[post]: %s = %v", key, value)
			}
			fieldFuncs := postConfig.GetFieldFuncs()
			for key, fn := range fieldFuncs {
				value := fn(item)
				setFieldAny(item, key, value)
				debugger.Pipeline("GlobalConfigFunc[post]: %s = %v", key, value)
			}
		}

		if p.customFields != nil {
			for key, value := range p.customFields {
				setFieldAny(item, key, value)
				debugger.Pipeline("SetFields: %s = %v", key, value)
			}
		}
	}

	resVal := reflect.ValueOf(res)
	if resVal.Kind() == reflect.Slice || resVal.Kind() == reflect.Array {
		for i := 0; i < resVal.Len(); i++ {
			item := resVal.Index(i).Addr().Interface()
			applyFields(item, "array")
		}
	} else {
		applyFields(&res, "single")
	}

	debugger.Pipeline("Respond success")
	return response.New[Res](p.ctx, p.status).Data(res).Send()
}

func (p *Pipeline[Req, Params, Entity, Res]) respondError() error {
	debugger.Error(debugger.ComponentPipeline, "📚 Reason : %v", p.err)
	if appErr, ok := p.err.(errors.Error); ok {
		return errors.New().Skip(3).
			Code(appErr.Details.Code).
			Summary("Request Failed").
			Detail(p.err).
			Send(p.ctx)
	}
	return errors.New().Skip(3).BadRequest().Summary("Request Failed").Detail(p.err).Send(p.ctx)
}

func (p *Pipeline[Req, Params, Entity, Res]) ensureID() {
	if p.err != nil {
		return
	}
	if p.id == nil {
		p.id = request.Param(p.ctx, "id")
	}
	if s, ok := p.id.(string); ok && s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			p.id = n
			debugger.Pipeline("autoParseID: %q → %d", s, n)
		} else {
			p.err = errors.New().Skip(2).BadRequest().
				Summary("Invalid ID format").
				Detail(fmt.Errorf("ID must be a number, got: %s", s)).
				Build()
			debugger.Error(debugger.ComponentPipeline, "Invalid ID format: %s", s)
		}
	}
}

func Validate() *validator.Validator { return validator.New() }

func setFieldAny(item any, fieldName string, value any) {
	itemVal := reflect.ValueOf(item)
	if itemVal.Kind() != reflect.Ptr || itemVal.IsNil() {
		return
	}
	elem := itemVal.Elem()
	if elem.Kind() != reflect.Struct {
		return
	}
	field := elem.FieldByName(fieldName)
	if !field.IsValid() || !field.CanSet() {
		return
	}
	fieldVal := reflect.ValueOf(value)
	if fieldVal.Type().AssignableTo(field.Type()) {
		field.Set(fieldVal)
	} else if field.Kind() == reflect.Int64 {
		switch v := value.(type) {
		case int64:
			field.SetInt(v)
		case int:
			field.SetInt(int64(v))
		case float64:
			field.SetInt(int64(v))
		}
	} else if field.Kind() == reflect.Int {
		switch v := value.(type) {
		case int64:
			field.SetInt(v)
		case int:
			field.SetInt(int64(v))
		case float64:
			field.SetInt(int64(v))
		}
	}
}
