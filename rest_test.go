/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package rest

import (
	"bytes"
	stdc "context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	restcontext "github.com/otmc-sw/rest/context"
)

type fakeContext struct {
	ctx    stdc.Context
	params map[string]string
	body   []byte
	status int
	wrote  interface{}
	header http.Header
}

func (f *fakeContext) Context() stdc.Context      { return f.ctx }
func (f *fakeContext) Param(k string) string      { return f.params[k] }
func (f *fakeContext) Query(k string) string      { return "" }
func (f *fakeContext) QueryAll(k string) []string { return nil }
func (f *fakeContext) Header(k string) string     { return f.header.Get(k) }
func (f *fakeContext) Cookie(k string) string     { return "" }
func (f *fakeContext) Body() io.Reader            { return bytes.NewReader(f.body) }
func (f *fakeContext) Bind(v interface{}) error   { return json.Unmarshal(f.body, v) }
func (f *fakeContext) JSON(code int, body interface{}) error {
	f.status = code
	f.wrote = body
	return nil
}
func (f *fakeContext) Status(code int)         { f.status = code }
func (f *fakeContext) SetHeader(k, v string)   { f.header.Set(k, v) }
func (f *fakeContext) Method() string          { return http.MethodPost }
func (f *fakeContext) Path() string            { return "/docs" }
func (f *fakeContext) String() (string, error) { return string(f.body), nil }
func (f *fakeContext) Bytes() ([]byte, error)  { return f.body, nil }

var _ restcontext.Context = (*fakeContext)(nil)

func newFakeContext() *fakeContext {
	return &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		header: http.Header{},
	}
}

type CreateDocRequest struct {
	Title string `json:"title"`
}

type DocEntity struct {
	ID    string
	Title string
}

type DocResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func TestCreatePipeline(t *testing.T) {
	p := Create[CreateDocRequest, struct{}, DocEntity, DocResponse](newFakeContext())
	if p == nil {
		t.Fatal("Create returned nil pipeline")
	}
}

func TestGetPipeline(t *testing.T) {
	p := Get[CreateDocRequest, struct{}, DocEntity, DocResponse](newFakeContext())
	if p == nil {
		t.Fatal("Get returned nil pipeline")
	}
}

func TestUpdatePipeline(t *testing.T) {
	p := Update[CreateDocRequest, struct{}, DocEntity, DocResponse](newFakeContext())
	if p == nil {
		t.Fatal("Update returned nil pipeline")
	}
}

func TestDeletePipeline(t *testing.T) {
	p := Delete[struct{}, struct{}, struct{}, DocResponse](newFakeContext())
	if p == nil {
		t.Fatal("Delete returned nil pipeline")
	}
}

func TestCreatePipelineEndToEnd(t *testing.T) {
	payload, _ := json.Marshal(CreateDocRequest{Title: "Hello"})
	fc := newFakeContext()
	fc.body = payload
	fc.params["id"] = "42"

	handle := func(ctx Context, req CreateDocRequest, id any) (DocEntity, error) {
		return DocEntity{ID: "42", Title: req.Title}, nil
	}

	err := Create[CreateDocRequest, struct{}, DocEntity, DocResponse](fc).
		Param("id").
		Bind().
		Validate(func(r CreateDocRequest) error {
			if r.Title == "" {
				return stdc.DeadlineExceeded
			}
			return nil
		}).
		Handle(handle).
		Respond()
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	if fc.status != 201 {
		t.Fatalf("expected status 201, got %d", fc.status)
	}

	data, _ := json.Marshal(fc.wrote)
	if string(data) == "" {
		t.Fatalf("expected a written body, got empty")
	}
}

func TestGetPipelineEndToEnd(t *testing.T) {
	fc := newFakeContext()
	fc.body = []byte("{}")
	fc.params["id"] = "7"

	handle := func(ctx Context, req CreateDocRequest, id any) (DocEntity, error) {
		return DocEntity{ID: "7", Title: "Fetched"}, nil
	}

	err := Get[CreateDocRequest, struct{}, DocEntity, DocResponse](fc).
		Param("id").
		Bind().
		Handle(handle).
		Respond()
	if err != nil {
		t.Fatalf("get pipeline failed: %v", err)
	}

	if fc.status != 200 {
		t.Fatalf("expected status 200, got %d", fc.status)
	}

	resp, ok := fc.wrote.(map[string]interface{})
	if !ok {
		if data, _ := json.Marshal(fc.wrote); string(data) == "" {
			t.Fatalf("expected a written body, got empty")
		}
		return
	}
	_ = resp
}

func TestDeletePipelineEndToEnd(t *testing.T) {
	fc := newFakeContext()
	fc.body = []byte("{}")
	fc.params["id"] = "99"

	handle := func(ctx Context, req struct{}, id any) (struct{}, error) {
		return struct{}{}, nil
	}

	err := Delete[struct{}, struct{}, struct{}, DocResponse](fc).
		Param("id").
		Bind().
		Handle(handle).
		Respond()
	if err != nil {
		t.Fatalf("delete pipeline failed: %v", err)
	}

	if fc.status != 204 {
		t.Fatalf("expected status 204, got %d", fc.status)
	}
}

func TestRegisterAndMap(t *testing.T) {
	Register(func(e DocEntity) DocResponse {
		return DocResponse{ID: e.ID, Title: e.Title}
	})

	fc := newFakeContext()
	fc.body, _ = json.Marshal(CreateDocRequest{Title: "Mapped"})
	fc.params["id"] = "55"

	handle := func(ctx Context, req CreateDocRequest, id any) (DocEntity, error) {
		return DocEntity{ID: "55", Title: req.Title}, nil
	}

	err := Create[CreateDocRequest, struct{}, DocEntity, DocResponse](fc).
		Param("id").
		Bind().
		Handle(handle).
		Respond()
	if err != nil {
		t.Fatalf("pipeline with registered mapper failed: %v", err)
	}

	data, _ := json.Marshal(fc.wrote)
	if string(data) == "" {
		t.Fatalf("expected a written body, got empty")
	}
}

func TestValidate(t *testing.T) {
	v := Validator()
	if v == nil {
		t.Fatal("Validate returned nil")
	}

	v.Required("value").Min("hello", 3).Email("a@b.com")
	if v.HasErrors() {
		t.Fatalf("expected no validation errors, got: %v", v.Errors())
	}

	v2 := Validator()
	v2.Required("").Email("not-an-email")
	if !v2.HasErrors() {
		t.Fatal("expected validation errors, got none")
	}
	if err := v2.Process(); err == nil {
		t.Fatal("expected Process() to return an error")
	}
}

func TestNewError(t *testing.T) {
	b := NewError()
	if b == nil {
		t.Fatal("NewError returned nil")
	}

	err := b.BadRequest().Summary("missing field").Detail("title is required").Build()
	if err.Details.Code != 400 {
		t.Fatalf("expected code 400, got %d", err.Details.Code)
	}
	if err.Details.Key != "BAD_REQUEST" {
		t.Fatalf("expected key BAD_REQUEST, got %s", err.Details.Key)
	}
	if err.Details.Summary != "missing field" {
		t.Fatalf("unexpected summary: %s", err.Details.Summary)
	}
	if err.Success {
		t.Fatal("expected success to be false for an error")
	}
}

func TestNewErrorSendWithoutContext(t *testing.T) {
	b := NewError().InternalError().Summary("boom")
	err := b.Send(nil)
	if err == nil {
		t.Fatal("expected an error when sending with nil context")
	}
}

func TestDebugFunctions(t *testing.T) {
	Debug()
	DebugComponent("mapper")
	DebugComponent("all")

	t.Setenv("REST_DEBUG", "pipeline,validator")
	DebugWithEnv()

	os.Unsetenv("REST_DEBUG")
}
