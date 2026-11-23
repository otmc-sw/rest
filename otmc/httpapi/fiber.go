/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	restcontext "github.com/otmc-sw/rest/context"
	resterrors "github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/otmc/schema"
	"github.com/otmc-sw/rest/otmc/store"
	"github.com/otmc-sw/rest/pagination"
	"github.com/otmc-sw/rest/response"
)

// reservedQueryKeys are never treated as field filters.
var reservedQueryKeys = map[string]bool{
	"page": true, "size": true, "limit": true, "offset": true,
	"keyword": true, "q": true, "sort": true, "order": true,
	"ids": true,
}

// RegisterRoutes mounts the auto-generated CRUD endpoints for every repository.
// prefix is the API root, e.g. "/api" -> "/api/users".
func RegisterRoutes(app *fiber.App, prefix string, repos []*store.Repository) {
	prefix = strings.TrimSuffix(prefix, "/")
	for _, repo := range repos {
		mountResource(app, prefix, repo)
	}
}

func mountResource(app *fiber.App, prefix string, repo *store.Repository) {
	m := repo.Model()
	base := prefix + "/" + m.Route
	h := &handlers{repo: repo}

	app.Get(base+"/search", h.search) // before /:id so "search" wins
	app.Get(base, h.list)
	app.Get(base+"/:id", h.get)
	app.Post(base, h.create)
	app.Put(base+"/:id", h.update)
	app.Patch(base+"/:id", h.patch)
	app.Delete(base+"/:id", h.delete)
	app.Delete(base, h.deleteMany)
}

type handlers struct {
	repo *store.Repository
}

func (h *handlers) search(c *fiber.Ctx) error {
	return h.list(c)
}

func (h *handlers) list(c *fiber.Ctx) error {
	ctx := newFiberContext(c)
	q := store.Query{
		Page:    pagination.New(ctx).Page().Number,
		Size:    pagination.New(ctx).Page().Size,
		Keyword: firstNonEmpty(query(c, "keyword"), query(c, "q")),
		Sort:    query(c, "sort"),
		Order:   query(c, "order"),
		Filters: map[string]string{},
	}
	for key, val := range c.Queries() {
		if !reservedQueryKeys[key] && val != "" {
			q.Filters[key] = val
		}
	}
	result, err := h.repo.List(c.Context(), q)
	if err != nil {
		return h.fail(ctx, err)
	}
	data := map[string]any{
		"items":       sanitize(result.Items, h.repo.Model()),
		"total":       result.Total,
		"page":        result.Page,
		"size":        result.Size,
		"total_pages": result.TotalPages,
	}
	return response.OK[any](ctx).Data(data).Send()
}

func (h *handlers) get(c *fiber.Ctx) error {
	ctx := newFiberContext(c)
	id, err := h.parseID(c)
	if err != nil {
		return badRequest(ctx, err)
	}
	inst, err := h.repo.Get(c.Context(), id)
	if err != nil {
		return h.fail(ctx, err)
	}
	return response.OK[any](ctx).Data(sanitize(inst, h.repo.Model())).Send()
}

func (h *handlers) create(c *fiber.Ctx) error {
	ctx := newFiberContext(c)
	inst, err := h.repo.Create(c.Context(), c.Body())
	if err != nil {
		return h.fail(ctx, err)
	}
	return response.Created[any](ctx).Data(sanitize(inst, h.repo.Model())).Send()
}

func (h *handlers) update(c *fiber.Ctx) error {
	ctx := newFiberContext(c)
	id, err := h.parseID(c)
	if err != nil {
		return badRequest(ctx, err)
	}
	inst, err := h.repo.Update(c.Context(), id, c.Body())
	if err != nil {
		return h.fail(ctx, err)
	}
	return response.OK[any](ctx).Data(sanitize(inst, h.repo.Model())).Send()
}

func (h *handlers) patch(c *fiber.Ctx) error {
	ctx := newFiberContext(c)
	id, err := h.parseID(c)
	if err != nil {
		return badRequest(ctx, err)
	}
	inst, err := h.repo.Patch(c.Context(), id, c.Body())
	if err != nil {
		return h.fail(ctx, err)
	}
	return response.OK[any](ctx).Data(sanitize(inst, h.repo.Model())).Send()
}

func (h *handlers) delete(c *fiber.Ctx) error {
	ctx := newFiberContext(c)
	id, err := h.parseID(c)
	if err != nil {
		return badRequest(ctx, err)
	}
	if err := h.repo.Delete(c.Context(), id); err != nil {
		return h.fail(ctx, err)
	}
	return response.OK[any](ctx).Message("deleted").Send()
}

func (h *handlers) deleteMany(c *fiber.Ctx) error {
	ctx := newFiberContext(c)
	raw := query(c, "ids")
	if raw == "" {
		return badRequest(ctx, fmt.Errorf("ids query parameter is required"))
	}
	parts := strings.Split(raw, ",")
	ids := make([]any, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := h.repo.ParseID(p)
		if err != nil {
			return badRequest(ctx, err)
		}
		ids = append(ids, id)
	}
	n, err := h.repo.DeleteMany(c.Context(), ids)
	if err != nil {
		return h.fail(ctx, err)
	}
	return response.OK[any](ctx).Data(map[string]any{"deleted": n}).Send()
}

func (h *handlers) parseID(c *fiber.Ctx) (any, error) {
	return h.repo.ParseID(c.Params("id"))
}

// fail maps store errors to HTTP error responses.
func (h *handlers) fail(ctx restcontext.Context, err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return resterrors.New().NotFound().Summary("record not found").Detail(err.Error()).Send(ctx)
	case isValidation(err):
		return resterrors.New().UnprocessableEntity().Summary("validation failed").Detail(err.Error()).Send(ctx)
	case isUnique(err):
		return resterrors.New().Conflict().Summary("conflict").Detail(err.Error()).Send(ctx)
	default:
		return resterrors.New().InternalError().Summary("internal server error").Detail(err.Error()).Send(ctx)
	}
}

func isValidation(err error) bool {
	var ve *store.ValidationError
	return errors.As(err, &ve)
}

func isUnique(err error) bool {
	var ue *store.UniqueError
	return errors.As(err, &ue)
}

func badRequest(ctx restcontext.Context, err error) error {
	return resterrors.New().BadRequest().Summary("bad request").Detail(err.Error()).Send(ctx)
}

func query(c *fiber.Ctx, key string) string {
	return c.Query(key)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// sanitize strips hidden fields from API output.
func sanitize(v any, m *schema.Model) any {
	hidden := map[string]bool{}
	for _, f := range m.Fields {
		if f.Hidden {
			hidden[f.JSONName] = true
		}
	}
	if len(hidden) == 0 {
		return v
	}
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return v
	}
	stripHidden(out, hidden)
	return out
}

func stripHidden(node any, hidden map[string]bool) {
	switch n := node.(type) {
	case map[string]any:
		for k, v := range n {
			if hidden[k] {
				delete(n, k)
			} else {
				stripHidden(v, hidden)
			}
		}
	case []any:
		for _, e := range n {
			stripHidden(e, hidden)
		}
	}
}

// ---------------------------------------------------------------------------
// FiberContext adapts *fiber.Ctx to restcontext.Context.
// ---------------------------------------------------------------------------

type FiberContext struct {
	*fiber.Ctx
}

func newFiberContext(c *fiber.Ctx) restcontext.Context {
	return FiberContext{Ctx: c}
}

func (c FiberContext) Context() context.Context { return c.Ctx.Context() }
func (c FiberContext) Param(key string) string  { return c.Ctx.Params(key) }
func (c FiberContext) Query(key string) string  { return c.Ctx.Query(key) }
func (c FiberContext) QueryAll(key string) []string {
	v := c.Ctx.Query(key)
	if v == "" {
		return nil
	}
	return []string{v}
}
func (c FiberContext) Header(key string) string { return c.Ctx.Get(key) }
func (c FiberContext) Cookie(key string) string { return c.Ctx.Cookies(key) }
func (c FiberContext) Body() io.Reader          { return bytes.NewReader(c.Ctx.Body()) }
func (c FiberContext) Bind(v interface{}) error { return c.Ctx.BodyParser(v) }
func (c FiberContext) JSON(code int, body interface{}) error {
	return c.Ctx.Status(code).JSON(body)
}
func (c FiberContext) Status(code int)             { c.Ctx.Status(code) }
func (c FiberContext) SetHeader(key, value string) { c.Ctx.Set(key, value) }
func (c FiberContext) Method() string              { return c.Ctx.Method() }
func (c FiberContext) Path() string                { return c.Ctx.Path() }
func (c FiberContext) String() (string, error)     { return string(c.Ctx.Body()), nil }
func (c FiberContext) Bytes() ([]byte, error)      { return c.Ctx.Body(), nil }
func (c FiberContext) SendFile(path string) error  { return c.Ctx.SendFile(path) }
func (c FiberContext) Download(path string, name string) error {
	return c.Ctx.Download(path, name)
}
func (c FiberContext) HTML(html string) error { return c.Ctx.Type("html").SendString(html) }
func (c FiberContext) Text(text string) error { return c.Ctx.Type("text").SendString(text) }
func (c FiberContext) Redirect(location string) error {
	return c.Ctx.Redirect(location)
}
func (c FiberContext) Stream(reader io.Reader) error {
	return c.Ctx.SendStream(reader)
}
func (c FiberContext) FormFile(key string) (interface{}, error) {
	return c.Ctx.FormFile(key)
}

// ParseUint64 is a small helper kept for callers that need raw id parsing.
func ParseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
