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
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	restcontext "github.com/otmc-sw/rest/context"
)

type fakeMultipartFile struct {
	filename string
	size     int64
	reader   io.ReadSeeker
}

func (f *fakeMultipartFile) Filename() string            { return f.filename }
func (f *fakeMultipartFile) Size() int64                 { return f.size }
func (f *fakeMultipartFile) Open() (io.ReadCloser, error) { return io.NopCloser(f.reader), nil }

type fakeContext struct {
	ctx      stdc.Context
	params   map[string]string
	body     []byte
	status   int
	wrote    interface{}
	header   http.Header
	formFile *fakeMultipartFile
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
	if f.formFile != nil {
		return f.formFile, nil
	}
	return nil, fmt.Errorf("no file uploaded")
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

func TestDownloadPipeline(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	content := []byte("Hello, World!")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte{},
		header: http.Header{},
	}

	var file File
	beforeCalled := false
	afterCalled := false

	err := Download(fc).
		Source(tmpFile).
		Bind(&file).
		Before(func(ctx restcontext.Context, f *File) error {
			beforeCalled = true
			if f.Name != "test.txt" {
				t.Fatalf("expected name test.txt, got %s", f.Name)
			}
			return nil
		}).
		After(func(ctx restcontext.Context, f *File) error {
			afterCalled = true
			if f.Size != int64(len(content)) {
				t.Fatalf("expected size %d, got %d", len(content), f.Size)
			}
			return nil
		}).
		Respond()

	if err != nil {
		t.Fatalf("download pipeline failed: %v", err)
	}

	if !beforeCalled {
		t.Fatal("expected before hook to be called")
	}
	if !afterCalled {
		t.Fatal("expected after hook to be called")
	}
	if file.Name != "test.txt" {
		t.Fatalf("expected bound file name test.txt, got %s", file.Name)
	}
	if file.Size != int64(len(content)) {
		t.Fatalf("expected bound file size %d, got %d", len(content), file.Size)
	}
	if file.ContentType == "" {
		t.Fatal("expected content type to be populated")
	}
	if file.ETag == "" {
		t.Fatal("expected etag to be populated")
	}
	if file.Extension != "txt" {
		t.Fatalf("expected extension txt, got %s", file.Extension)
	}
}

func TestDownloadPipelineFileNotFound(t *testing.T) {
	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte{},
		header: http.Header{},
	}

	Download(fc).
		Source("/nonexistent/file.txt").
		Respond()

	if fc.status != 404 {
		t.Fatalf("expected status 404, got %d", fc.status)
	}

	if fc.wrote == nil {
		t.Fatal("expected error response to be written")
	}
}

func TestDownloadPipelineBeforeAborts(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte{},
		header: http.Header{},
	}

	Download(fc).
		Source(tmpFile).
		Before(func(ctx restcontext.Context, f *File) error {
			return fmt.Errorf("aborted")
		}).
		Respond()

	if fc.wrote == nil {
		t.Fatal("expected error response to be written")
	}
}

func TestUploadPipeline(t *testing.T) {
	destDir := t.TempDir()
	content := []byte("uploaded file content")

	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   content,
		header: http.Header{},
		formFile: &fakeMultipartFile{
			filename: "upload.txt",
			size:     int64(len(content)),
			reader:   bytes.NewReader(content),
		},
	}

	var file File
	beforeCalled := false
	afterCalled := false

	err := Upload(fc).
		Destination(destDir).
		Bind(&file).
		Before(func(ctx restcontext.Context, f *File) error {
			beforeCalled = true
			if f.Name != "upload.txt" {
				t.Fatalf("expected name upload.txt, got %s", f.Name)
			}
			return nil
		}).
		After(func(ctx restcontext.Context, f *File) error {
			afterCalled = true
			if f.Size != int64(len(content)) {
				t.Fatalf("expected size %d, got %d", len(content), f.Size)
			}
			return nil
		}).
		Respond()

	if err != nil {
		t.Fatalf("upload pipeline failed: %v", err)
	}

	if !beforeCalled {
		t.Fatal("expected before hook to be called")
	}
	if !afterCalled {
		t.Fatal("expected after hook to be called")
	}
	if file.Name != "upload.txt" {
		t.Fatalf("expected bound file name upload.txt, got %s", file.Name)
	}
	if file.Size != int64(len(content)) {
		t.Fatalf("expected bound file size %d, got %d", len(content), file.Size)
	}
	if file.Extension != "txt" {
		t.Fatalf("expected extension txt, got %s", file.Extension)
	}

	savedPath := filepath.Join(destDir, "upload.txt")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Fatal("expected uploaded file to exist on disk")
	}
	savedContent, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(savedContent) != string(content) {
		t.Fatalf("expected saved content %q, got %q", content, savedContent)
	}
}

func TestUploadPipelineNoDestination(t *testing.T) {
	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte{},
		header: http.Header{},
	}

	Upload(fc).Respond()

	if fc.wrote == nil {
		t.Fatal("expected error response to be written")
	}
}

func TestUploadPipelineBeforeAborts(t *testing.T) {
	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte("data"),
		header: http.Header{},
		formFile: &fakeMultipartFile{
			filename: "test.txt",
			size:     4,
			reader:   bytes.NewReader([]byte("data")),
		},
	}

	Upload(fc).
		Destination(t.TempDir()).
		Before(func(ctx restcontext.Context, f *File) error {
			return fmt.Errorf("aborted")
		}).
		Respond()

	if fc.wrote == nil {
		t.Fatal("expected error response to be written")
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

func TestPreviewFileContentPipeline(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "preview.txt")
	content := []byte("preview test content")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fc := &fakeContext{
		ctx:    stdc.Background(),
		params: map[string]string{},
		body:   []byte{},
		header: http.Header{},
	}

	var fileContent FileContent
	beforeCalled := false
	afterCalled := false

	err := PreviewFileContent(fc).
		Source(tmpFile).
		Bind(&fileContent).
		Before(func(ctx restcontext.Context, fc *FileContent) error {
			beforeCalled = true
			if fc.Content != string(content) {
				t.Fatalf("expected content %s, got %s", content, fc.Content)
			}
			return nil
		}).
		After(func(ctx restcontext.Context, fc *FileContent) error {
			afterCalled = true
			if fc.Size != int64(len(content)) {
				t.Fatalf("expected size %d, got %d", len(content), fc.Size)
			}
			return nil
		}).
		Respond()

	if err != nil {
		t.Fatalf("preview file content pipeline failed: %v", err)
	}

	if !beforeCalled {
		t.Fatal("expected before hook to be called")
	}
	if !afterCalled {
		t.Fatal("expected after hook to be called")
	}
	if fileContent.Content != string(content) {
		t.Fatalf("expected bound content %s, got %s", content, fileContent.Content)
	}
	if fc.wrote != tmpFile {
		t.Fatalf("expected SendFile with %s, got %v", tmpFile, fc.wrote)
	}
}

