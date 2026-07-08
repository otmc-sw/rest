/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
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

func (f *fakeContext) Context() stdc.Context { return f.ctx }
func (f *fakeContext) Param(k string) string { return f.params[k] }
func (f *fakeContext) Query(k string) string { return "" }
func (f *fakeContext) QueryAll(k string) []string { return nil }
func (f *fakeContext) Header(k string) string { return f.header.Get(k) }
func (f *fakeContext) Cookie(k string) string { return "" }
func (f *fakeContext) Body() io.Reader { return bytes.NewReader(f.body) }
func (f *fakeContext) Bind(v interface{}) error { return json.Unmarshal(f.body, v) }
func (f *fakeContext) JSON(code int, body interface{}) error {
	f.status = code
	f.wrote = body
	return nil
}
func (f *fakeContext) Status(code int) { f.status = code }
func (f *fakeContext) SetHeader(k, v string) { f.header.Set(k, v) }
func (f *fakeContext) Method() string { return http.MethodPost }
func (f *fakeContext) Path() string { return "/docs" }
func (f *fakeContext) String() (string, error) { return string(f.body), nil }
func (f *fakeContext) Bytes() ([]byte, error) { return f.body, nil }

var _ restcontext.Context = (*fakeContext)(nil)

type CreateDocRequest struct {
	Title string `json:"title"`
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

	handle := func(ctx restcontext.Context, req CreateDocRequest) (DocResponse, error) {
		return DocResponse{ID: "42", Title: req.Title}, nil
	}

	err := Create[CreateDocRequest, DocResponse](fc).
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

	if fc.status != 200 {
		t.Fatalf("expected status 200, got %d", fc.status)
	}
}