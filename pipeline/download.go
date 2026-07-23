/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package pipeline

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/debugger"
	"github.com/otmc-sw/rest/errors"
)

type File struct {
	Name         string            `json:"name"`
	OriginalName string            `json:"original_name"`
	Extension    string            `json:"extension"`
	Path         string            `json:"path"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	Reader       io.ReadSeeker     `json:"-"`
	LastModified time.Time         `json:"last_modified"`
	ETag         string            `json:"etag"`
	Metadata     map[string]any    `json:"metadata,omitempty"`
}

type DownloadPipeline struct {
	ctx     context.Context
	source  string
	file    *File
	before  func(context.Context, *File) error
	after   func(context.Context, *File) error
	err     error
}

func Download(ctx context.Context) *DownloadPipeline {
	debugger.Pipeline("Download")
	return &DownloadPipeline{
		ctx: ctx,
	}
}

func (p *DownloadPipeline) Source(path string) *DownloadPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Download.Source", "path=%s", path)
	p.source = path
	return p
}

func (p *DownloadPipeline) Bind(f *File) *DownloadPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Download.Bind", "binding file")
	p.file = f
	return p
}

func (p *DownloadPipeline) Before(fn func(context.Context, *File) error) *DownloadPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Download.Before", "registering before hook")
	p.before = fn
	return p
}

func (p *DownloadPipeline) After(fn func(context.Context, *File) error) *DownloadPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("Download.After", "registering after hook")
	p.after = fn
	return p
}

func (p *DownloadPipeline) Respond() error {
	debugger.PipelineStep("Download.Respond", "executing download pipeline")

	if p.err != nil {
		return p.respondError()
	}

	if p.source == "" {
		err := errors.New().Skip(2).BadRequest().Summary("source path is required").Build()
		return err.Err()
	}

	debugger.PipelineStep("Download", "opening file: %s", p.source)
	fileInfo, err := os.Stat(p.source)
	if err != nil {
		if os.IsNotExist(err) {
			debugger.Error(debugger.ComponentPipeline, "file not found: %s", p.source)
			return errors.New().Skip(2).NotFound().
				Summary("File not found").
				Detail(fmt.Sprintf("The requested file does not exist: %s", p.source)).
				Send(p.ctx)
		}
		debugger.Error(debugger.ComponentPipeline, "failed to stat file: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to access file").
			Detail(err.Error()).
			Send(p.ctx)
	}

	f, err := os.Open(p.source)
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to open file: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to open file").
			Detail(err.Error()).
			Send(p.ctx)
	}
	defer f.Close()

	originalName := filepath.Base(p.source)
	ext := strings.TrimPrefix(filepath.Ext(originalName), ".")
	contentType := mime.TypeByExtension(filepath.Ext(originalName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to compute hash: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to compute file hash").
			Detail(err.Error()).
			Send(p.ctx)
	}
	etag := fmt.Sprintf("\"%x\"", hash.Sum(nil)[:16])

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to seek file: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to read file").
			Detail(err.Error()).
			Send(p.ctx)
	}

	file := &File{
		Name:         originalName,
		OriginalName: originalName,
		Extension:    ext,
		Path:         p.source,
		Size:         fileInfo.Size(),
		ContentType:  contentType,
		Reader:       f,
		LastModified: fileInfo.ModTime(),
		ETag:         etag,
		Metadata:     make(map[string]any),
	}

	if p.before != nil {
		debugger.PipelineStep("Download", "executing before hook")
		if err := p.before(p.ctx, file); err != nil {
			debugger.Error(debugger.ComponentPipeline, "before hook failed: %v", err)
			return errors.New().Skip(2).InternalError().
				Summary("Download aborted by before hook").
				Detail(err.Error()).
				Send(p.ctx)
		}
	}

	if p.after != nil {
		debugger.PipelineStep("Download", "executing after hook")
		if err := p.after(p.ctx, file); err != nil {
			debugger.Error(debugger.ComponentPipeline, "after hook failed: %v", err)
			return errors.New().Skip(2).InternalError().
				Summary("Download after hook failed").
				Detail(err.Error()).
				Send(p.ctx)
		}
	}

	if p.file != nil {
		*p.file = *file
	}

	p.ctx.SetHeader("Content-Type", contentType)
	p.ctx.SetHeader("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	p.ctx.SetHeader("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, originalName))
	p.ctx.SetHeader("Last-Modified", fileInfo.ModTime().UTC().Format(httpTimeFormat))
	p.ctx.SetHeader("ETag", etag)
	p.ctx.SetHeader("Accept-Ranges", "bytes")

	debugger.PipelineStep("Download", "sending file: %s", p.source)
	if err := p.ctx.SendFile(p.source); err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to send file: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to send file").
			Detail(err.Error()).
			Send(p.ctx)
	}

	debugger.Pipeline("Download success: %s (%d bytes)", originalName, fileInfo.Size())
	return nil
}

func (p *DownloadPipeline) respondError() error {
	debugger.Error(debugger.ComponentPipeline, "Download error: %v", p.err)
	if appErr, ok := p.err.(errors.Error); ok {
		return errors.New().Skip(3).
			Code(appErr.Details.Code).
			Summary("Download Failed").
			Detail(p.err).
			Send(p.ctx)
	}
	return errors.New().Skip(3).BadRequest().Summary("Download Failed").Detail(p.err).Send(p.ctx)
}

const httpTimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

var _ = Download