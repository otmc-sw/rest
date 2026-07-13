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

// Handler dùng chung cho Create/Update khi không cần tách biệt Params
type Handler[Req any, Entity any] func(ctx context.Context, req Req, id any) (Entity, error)

// ExecHandler dùng chung cho các thao tác xử lý nghiệp vụ trả về any
type ExecHandler[Req any] func(ctx context.Context, req Req, id any) (any, error)

// PatchHandler dùng cho trường hợp cần tách biệt Request DTO và Params DTO
type PatchHandler[Req any, Params any] func(ctx context.Context, req Req, params Params, id any) (any, error)

// Pipeline là pipeline duy nhất dùng chung cho Create, Update, Patch
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

// Create khởi tạo pipeline với status 201.
// Nếu không cần Params riêng biệt, hãy set Params = Req hoặc struct{}
func Create[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Create[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 201)
}

// Get vẫn giữ nguyên nếu bạn còn dùng, hoặc có thể chuyển sang Pipeline tương tự
func Get[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Get[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200)
}

// Update khởi tạo pipeline với status 200
func Update[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Update[%T, %T] start", *new(Req), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200)
}

// Delete khởi tạo pipeline với status 204
func Delete[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Delete start")
	return newPipeline[Req, Params, Entity, Res](ctx, 204)
}

// Patch khởi tạo pipeline với status 200
func Patch[Req any, Params any, Entity any, Res any](ctx context.Context) *Pipeline[Req, Params, Entity, Res] {
	debugger.Pipeline("Patch[%T, %T, %T] start", *new(Req), *new(Params), *new(Entity))
	return newPipeline[Req, Params, Entity, Res](ctx, 200)
}

// Param lấy ID từ URL path parameter
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

// Params xây dựng đối tượng Params từ Request (dành riêng cho Patch/Update phức tạp)
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

// Handle hỗ trợ handler signature cũ (không có params) để tương thích ngược với Create/Update đơn giản
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

// Exec hỗ trợ handler signature đầy đủ (có params)
func (p *Pipeline[Req, Params, Entity, Res]) Exec(handler PatchHandler[Req, Params]) *Pipeline[Req, Params, Entity, Res] {
	if p.bindErr != nil {
		return p
	}
	if p.bound == nil {
		var req Req
		p.bound = &req
	}
	p.ensureID()
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
		// Fallback: Ưu tiên map từ params nếu có, ngược lại map từ bound request
		// Điều này giúp Create (thường không có params) vẫn hoạt động đúng
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
