/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package pipeline

import (
	"bytes"
	stdc "context"
	"encoding/json"
	"io"
	"net/http"
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
func (f *fakeContext) SendFile(path string) error {
	f.wrote = path
	return nil
}
func (f *fakeContext) Download(path string, name string) error {
	f.wrote = map[string]string{"path": path, "name": name}
	return nil
}
func (f *fakeContext) HTML(html string) error {
	f.wrote = html
	return nil
}
func (f *fakeContext) Text(text string) error {
	f.wrote = text
	return nil
}
func (f *fakeContext) Redirect(location string) error {
	f.wrote = location
	f.status = 302
	return nil
}
func (f *fakeContext) Stream(reader io.Reader) error {
	f.wrote = reader
	return nil
}
func (f *fakeContext) FormFile(key string) (interface{}, error) {
	return nil, nil
}

var _ restcontext.Context = (*fakeContext)(nil)

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

func TestPipelineEndToEnd(t *testing.T) {
	payload, _ := json.Marshal(CreateDocRequest{Title: "Hello"})
	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{"id": "42"},
		body:   payload,
		header: http.Header{},
	}

	handle := func(ctx restcontext.Context, req CreateDocRequest, id any) (DocEntity, error) {
		return DocEntity{ID: "42", Title: req.Title}, nil
	}

	err := Create[CreateDocRequest, CreateDocRequest, DocEntity, DocResponse](fc).
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

	resp := fc.wrote
	data, _ := json.Marshal(resp)
	if string(data) == "" {
		t.Fatalf("expected a written body, got empty")
	}
}

func TestInvalidIDFormat(t *testing.T) {
	payload, _ := json.Marshal(CreateDocRequest{Title: "Hello"})
	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{"id": "abc"},
		body:   payload,
		header: http.Header{},
	}

	handle := func(ctx restcontext.Context, req CreateDocRequest, id any) (DocEntity, error) {
		return DocEntity{ID: "abc", Title: req.Title}, nil
	}

	Create[CreateDocRequest, CreateDocRequest, DocEntity, DocResponse](fc).
		Param("id").
		Bind().
		Handle(handle).
		Respond()

	if fc.status != 400 {
		t.Fatalf("expected status 400 for invalid ID format, got %d", fc.status)
	}

	if fc.wrote == nil {
		t.Fatal("expected error response to be written")
	}
}

func TestRawPipeline(t *testing.T) {
	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte{},
		header: http.Header{},
	}

	err := Raw(fc).
		Exec(func(ctx restcontext.Context) error {
			return ctx.SendFile("/tmp/test.pdf")
		}).
		Respond()
	if err != nil {
		t.Fatalf("raw pipeline failed: %v", err)
	}

	if fc.wrote != "/tmp/test.pdf" {
		t.Fatalf("expected wrote to be /tmp/test.pdf, got %v", fc.wrote)
	}
}

func TestRawPipelineWithMiddleware(t *testing.T) {
	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte{},
		header: http.Header{},
	}

	middlewareCalled := false

	err := Raw(fc).
		Middleware(func(ctx restcontext.Context, next func(restcontext.Context) error) error {
			middlewareCalled = true
			return next(ctx)
		}).
		Exec(func(ctx restcontext.Context) error {
			return ctx.HTML("<h1>Hello</h1>")
		}).
		Respond()
	if err != nil {
		t.Fatalf("raw pipeline with middleware failed: %v", err)
	}

	if !middlewareCalled {
		t.Fatal("expected middleware to be called")
	}

	if fc.wrote != "<h1>Hello</h1>" {
		t.Fatalf("expected wrote to be HTML, got %v", fc.wrote)
	}
}

func TestRawPipelineNoHandler(t *testing.T) {
	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte{},
		header: http.Header{},
	}

	err := Raw(fc).Respond()
	if err != nil {
		t.Fatalf("expected no error return from Respond when handler sends error, got: %v", err)
	}

	if fc.wrote == nil {
		t.Fatal("expected error response to be written to context")
	}
}
