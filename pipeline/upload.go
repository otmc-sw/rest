/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package pipeline

import (
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/debugger"
	"github.com/otmc-sw/rest/errors"
)

type UploadPipeline struct {
	ctx         context.Context
	destination string
	file        *File
	before      func(context.Context, *File) error
	after       func(context.Context, *File) error
	err         error
}

func Upload(ctx context.Context) *UploadPipeline {
	debugger.Pipeline("Upload")
	return &UploadPipeline{
		ctx: ctx,
	}
}

func (p *UploadPipeline) Destination(path string) *UploadPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Upload.Destination", "path=%s", path)
	p.destination = path
	return p
}

func (p *UploadPipeline) Bind(f *File) *UploadPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Upload.Bind", "binding file")
	p.file = f
	return p
}

func (p *UploadPipeline) Before(fn func(context.Context, *File) error) *UploadPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Upload.Before", "registering before hook")
	p.before = fn
	return p
}

func (p *UploadPipeline) After(fn func(context.Context, *File) error) *UploadPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Upload.After", "registering after hook")
	p.after = fn
	return p
}

func (p *UploadPipeline) Respond() error {
	debugger.PipelineStep("Upload.Respond", "executing upload pipeline")

	if p.err != nil {
		return p.respondError()
	}

	if p.destination == "" {
		return errors.New().Skip(2).BadRequest().Summary("destination path is required").Send(p.ctx)
	}

	uploadedFile, err := p.ctx.FormFile("file")
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to get uploaded file: %v", err)
		return errors.New().Skip(2).BadRequest().
			Summary("Failed to read uploaded file").
			Detail(err.Error()).
			Send(p.ctx)
	}

	fileHeader, ok := uploadedFile.(*multipart.FileHeader)
	if !ok {
		debugger.Error(debugger.ComponentPipeline, "unexpected FormFile type: %T", uploadedFile)
		return errors.New().Skip(2).InternalError().
			Summary("Unexpected file type returned from form").
			Send(p.ctx)
	}

	originalName := filepath.Base(fileHeader.Filename)
	ext := strings.TrimPrefix(filepath.Ext(originalName), ".")
	contentType := mime.TypeByExtension(filepath.Ext(originalName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	src, err := fileHeader.Open()
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to open uploaded file: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to open uploaded file").
			Detail(err.Error()).
			Send(p.ctx)
	}
	defer src.Close()

	file := &File{
		Name:         originalName,
		OriginalName: originalName,
		Extension:    ext,
		Path:         filepath.Join(p.destination, originalName),
		Size:         fileHeader.Size,
		ContentType:  contentType,
		LastModified: time.Now(),
		Metadata:     make(map[string]any),
	}

	if p.before != nil {
		debugger.PipelineStep("Upload", "executing before hook")
		if err := p.before(p.ctx, file); err != nil {
			debugger.Error(debugger.ComponentPipeline, "before hook failed: %v", err)
			return errors.New().Skip(2).InternalError().
				Summary("Upload aborted by before hook").
				Detail(err.Error()).
				Send(p.ctx)
		}
	}

	if err := os.MkdirAll(p.destination, 0755); err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to create destination directory: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to create destination directory").
			Detail(err.Error()).
			Send(p.ctx)
	}

	dst, err := os.Create(file.Path)
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to create destination file: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to save file").
			Detail(err.Error()).
			Send(p.ctx)
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to write file: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to write file").
			Detail(err.Error()).
			Send(p.ctx)
	}

	file.Size = written

	if fi, err := os.Stat(file.Path); err == nil {
		file.LastModified = fi.ModTime()
	}

	if p.after != nil {
		debugger.PipelineStep("Upload", "executing after hook")
		if err := p.after(p.ctx, file); err != nil {
			debugger.Error(debugger.ComponentPipeline, "after hook failed: %v", err)
			return errors.New().Skip(2).InternalError().
				Summary("Upload after hook failed").
				Detail(err.Error()).
				Send(p.ctx)
		}
	}

	if p.file != nil {
		*p.file = *file
	}

	debugger.PipelineStep("Upload", "returning JSON response")
	return p.ctx.JSON(201, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"name":          file.Name,
			"original_name": file.OriginalName,
			"extension":     file.Extension,
			"path":          file.Path,
			"size":          file.Size,
			"content_type":  file.ContentType,
		},
		"message": "File uploaded successfully",
	})
}

func (p *UploadPipeline) respondError() error {
	debugger.Error(debugger.ComponentPipeline, "Upload error: %v", p.err)
	if appErr, ok := p.err.(errors.Error); ok {
		return errors.New().Skip(3).
			Code(appErr.Details.Code).
			Summary("Upload Failed").
			Detail(p.err).
			Send(p.ctx)
	}
	return errors.New().Skip(3).BadRequest().Summary("Upload Failed").Detail(p.err).Send(p.ctx)
}

var _ = Upload