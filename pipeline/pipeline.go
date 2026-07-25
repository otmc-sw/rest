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

	"github.com/google/uuid"
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
type ExecHandler[Req any, Params any] func(ctx context.Context, req Req, params Params, id any) (any, error)

type RawHandler interface {
	Exec(fn func(context.Context) error) RawHandler
	Middleware(fn func(context.Context, func(context.Context) error) error) RawHandler
	Respond() error
}

type rawHandler struct {
	ctx         context.Context
	exec        func(context.Context) error
	middlewares []func(context.Context, func(context.Context) error) error
}

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
	autoUUID     bool
}

func newPipeline[Req any, Params any, Entity any, Res any](ctx context.Context, status int, op string) *Pipeline[Req, Params, Entity, Res] {
	return &Pipeline[Req, Params, Entity, Res]{
		ctx:       ctx,
		status:    status,
		operation: op,
	}
}

func Create[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Create[%T, %T]", *new(Req), *new(Entity))
	p := newPipeline[Req, Params, Entity, Res](ctx, 201, "Post")
	p.autoUUID = true
	return p
}

func Get[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Get[%T, %T]", *new(Req), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200, "Get")
}

func Update[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Update[%T, %T]", *new(Req), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200, "Update")
}

func Patch[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Patch[%T, %T, %T]", *new(Req), *new(Params), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200, "Patch")
}

func Delete[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Delete")
	return newPipeline[Req, Params, Entity, Res](ctx, 204, "Delete")
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

func (p *Pipeline[Req, Params, Entity, Res]) AutoUUID(enable bool) *Pipeline[Req, Params, Entity, Res] {
	p.autoUUID = enable
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

	if p.autoUUID {
		ensureEntityUUID(&entity)
	}

	p.entity = &entity
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Exec(h ExecHandler[Req, Params]) *Pipeline[Req, Params, Entity, Res] {
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

	var e Entity
	switch {
	case result != nil:
		e = mapper.Map[Entity](result)
	case p.paramsFn != nil:
		e = mapper.Map[Entity](p.params)
	default:
		e = mapper.Map[Entity](*p.bound)
	}

	if p.autoUUID {
		ensureEntityUUID(&e)
	}

	p.entity = &e
	return p
}

func (p *Pipeline[Req, Params, Entity, Res]) Respond() error {
	debugger.PipelineStep("Respond", "status=%d", p.status)

	if p.err != nil {
		return p.respondError()
	}
	if p.entityFn != nil {
		e := p.entityFn()
		if p.autoUUID {
			ensureEntityUUID(&e)
		}
		p.entity = &e
	}
	if p.entity == nil {
		debugger.Error(debugger.ComponentPipeline, "Respond: no entity produced")
		return errors.New().Skip(2).InternalError().Summary("no result produced").Send(p.ctx)
	}

	res := mapper.Map[Res](*p.entity)

	applyFields := func(item any) {
		globalCfg := config.GetGlobalConfig()
		if globalCfg != nil {
			configs := []interface {
				GetFields() map[string]any
				GetFieldFuncs() map[string]func(any) any
			}{globalCfg.Pre(), globalCfg.Post()}

			for _, cfg := range configs {
				if cfg == nil {
					continue
				}
				for k, v := range cfg.GetFields() {
					setFieldAny(item, k, v)
					debugger.Pipeline("GlobalConfig: %s = %v", k, v)
				}
				for k, fn := range cfg.GetFieldFuncs() {
					val := fn(item)
					setFieldAny(item, k, val)
					debugger.Pipeline("GlobalConfigFunc: %s = %v", k, val)
				}
			}
		}

		if p.customFields != nil {
			for k, v := range p.customFields {
				setFieldAny(item, k, v)
				debugger.Pipeline("SetFields: %s = %v", k, v)
			}
		}
	}

	resVal := reflect.ValueOf(res)
	if resVal.Kind() == reflect.Slice || resVal.Kind() == reflect.Array {
		for i := 0; i < resVal.Len(); i++ {
			item := resVal.Index(i).Addr().Interface()
			applyFields(item)
		}
	} else {
		applyFields(&res)
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
		}
	}
}

func Validate() *validator.Validator { return validator.New() }

func ensureEntityUUID(entity any) {
	val := reflect.ValueOf(entity)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}
	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return
	}

	field := elem.FieldByName("ID")
	if !field.IsValid() {
		field = elem.FieldByName("Id")
	}

	if field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
		if field.String() == "" {
			newUUID := uuid.NewString()
			field.SetString(newUUID)
			debugger.Pipeline("autoGenerateUUID: generated %s", newUUID)
		}
	}
}

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
	if fieldVal.IsValid() && fieldVal.Type().AssignableTo(field.Type()) {
		field.Set(fieldVal)
		return
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		switch v := value.(type) {
		case int64:
			field.SetInt(v)
		case int:
			field.SetInt(int64(v))
		case float64:
			field.SetInt(int64(v))
		case string:
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				field.SetInt(n)
			}
		}
	case reflect.String:
		if v, ok := value.(fmt.Stringer); ok {
			field.SetString(v.String())
		} else {
			field.SetString(fmt.Sprintf("%v", value))
		}
	}
}

func Raw(ctx context.Context) RawHandler {
	debugger.Pipeline("Raw")
	return &rawHandler{
		ctx:         ctx,
		middlewares: []func(context.Context, func(context.Context) error) error{},
	}
}

func (h *rawHandler) Exec(fn func(context.Context) error) RawHandler {
	debugger.PipelineStep("Raw.Exec", "setting handler")
	h.exec = fn
	return h
}

func (h *rawHandler) Middleware(fn func(context.Context, func(context.Context) error) error) RawHandler {
	debugger.PipelineStep("Raw.Middleware", "adding middleware")
	h.middlewares = append(h.middlewares, fn)
	return h
}

func (h *rawHandler) Respond() error {
	debugger.PipelineStep("Raw.Respond", "executing")

	if h.exec == nil {
		debugger.Error(debugger.ComponentPipeline, "Raw: no handler provided")
		return errors.New().Skip(2).InternalError().Summary("no handler provided").Send(h.ctx)
	}

	handler := h.exec

	for i := len(h.middlewares) - 1; i >= 0; i-- {
		mw := h.middlewares[i]
		next := handler
		handler = func(ctx context.Context) error {
			return mw(ctx, next)
		}
	}

	return handler(h.ctx)
}
