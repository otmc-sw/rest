/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package pipeline

import (
	"io"
	"os"

	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/debugger"
	"github.com/otmc-sw/rest/errors"
)

type FileContent struct {
	Content string `json:"content"`
	Size    int64  `json:"size"`
}

type ReadFileContentPipeline struct {
	ctx     context.Context
	source  string
	content *FileContent
	before  func(context.Context, *FileContent) error
	after   func(context.Context, *FileContent) error
	err     error
}

func ReadFileContent(ctx context.Context) *ReadFileContentPipeline {
	debugger.Pipeline("ReadFileContent")
	return &ReadFileContentPipeline{
		ctx: ctx,
	}
}

func (p *ReadFileContentPipeline) Source(path string) *ReadFileContentPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("ReadFileContent.Source", "path=%s", path)
	p.source = path
	return p
}

func (p *ReadFileContentPipeline) Bind(fc *FileContent) *ReadFileContentPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("ReadFileContent.Bind", "binding file content")
	p.content = fc
	return p
}

func (p *ReadFileContentPipeline) Before(fn func(context.Context, *FileContent) error) *ReadFileContentPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("ReadFileContent.Before", "registering before hook")
	p.before = fn
	return p
}

func (p *ReadFileContentPipeline) After(fn func(context.Context, *FileContent) error) *ReadFileContentPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("ReadFileContent.After", "registering after hook")
	p.after = fn
	return p
}

func (p *ReadFileContentPipeline) Respond() error {
	debugger.PipelineStep("ReadFileContent.Respond", "executing read file content pipeline")

	if p.err != nil {
		return p.respondError()
	}

	if p.source == "" {
		err := errors.New().Skip(2).BadRequest().Summary("source path is required").Build()
		return err.Err()
	}

	debugger.PipelineStep("ReadFileContent", "reading file: %s", p.source)
	fileInfo, err := os.Stat(p.source)
	if err != nil {
		if os.IsNotExist(err) {
			debugger.Error(debugger.ComponentPipeline, "file not found: %s", p.source)
			return errors.New().Skip(2).NotFound().
				Summary("File not found").
				Detail(err.Error()).
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

	content, err := io.ReadAll(f)
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to read file content: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to read file content").
			Detail(err.Error()).
			Send(p.ctx)
	}

	fileContent := &FileContent{
		Content: string(content),
		Size:    fileInfo.Size(),
	}

	if p.before != nil {
		debugger.PipelineStep("ReadFileContent", "executing before hook")
		if err := p.before(p.ctx, fileContent); err != nil {
			debugger.Error(debugger.ComponentPipeline, "before hook failed: %v", err)
			return errors.New().Skip(2).InternalError().
				Summary("Read aborted by before hook").
				Detail(err.Error()).
				Send(p.ctx)
		}
	}

	if p.after != nil {
		debugger.PipelineStep("ReadFileContent", "executing after hook")
		if err := p.after(p.ctx, fileContent); err != nil {
			debugger.Error(debugger.ComponentPipeline, "after hook failed: %v", err)
			return errors.New().Skip(2).InternalError().
				Summary("Read after hook failed").
				Detail(err.Error()).
				Send(p.ctx)
		}
	}

	if p.content != nil {
		*p.content = *fileContent
	}

	debugger.PipelineStep("ReadFileContent", "returning JSON response")
	return p.ctx.JSON(200, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"content": fileContent.Content,
			"size":    fileContent.Size,
		},
		"message": "File content read successfully",
	})
}

func (p *ReadFileContentPipeline) respondError() error {
	debugger.Error(debugger.ComponentPipeline, "ReadFileContent error: %v", p.err)
	if appErr, ok := p.err.(errors.Error); ok {
		return errors.New().Skip(3).
			Code(appErr.Details.Code).
			Summary("Read File Content Failed").
			Detail(p.err).
			Send(p.ctx)
	}
	return errors.New().Skip(3).BadRequest().Summary("Read File Content Failed").Detail(p.err).Send(p.ctx)
}

type UpdateFileContentPipeline struct {
	ctx     context.Context
	source  string
	content *FileContent
	before  func(context.Context, *FileContent) error
	after   func(context.Context, *FileContent) error
	err     error
}

func UpdateFileContent(ctx context.Context) *UpdateFileContentPipeline {
	debugger.Pipeline("UpdateFileContent")
	return &UpdateFileContentPipeline{
		ctx: ctx,
	}
}

func (p *UpdateFileContentPipeline) Source(path string) *UpdateFileContentPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("UpdateFileContent.Source", "path=%s", path)
	p.source = path
	return p
}

func (p *UpdateFileContentPipeline) Bind(fc *FileContent) *UpdateFileContentPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("UpdateFileContent.Bind", "binding file content")
	p.content = fc
	return p
}

func (p *UpdateFileContentPipeline) Before(fn func(context.Context, *FileContent) error) *UpdateFileContentPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("UpdateFileContent.Before", "registering before hook")
	p.before = fn
	return p
}

func (p *UpdateFileContentPipeline) After(fn func(context.Context, *FileContent) error) *UpdateFileContentPipeline {
	if p.err != nil {
		return p
	}
	debugger.PipelineStep("UpdateFileContent.After", "registering after hook")
	p.after = fn
	return p
}

func (p *UpdateFileContentPipeline) Respond() error {
	debugger.PipelineStep("UpdateFileContent.Respond", "executing update file content pipeline")

	if p.err != nil {
		return p.respondError()
	}

	if p.source == "" {
		return errors.New().Skip(2).BadRequest().Summary("source path is required").Send(p.ctx)
	}

	var newContent string
	if p.content != nil {
		newContent = p.content.Content
	} else {
		var req struct {
			Content string `json:"content"`
		}
		if err := p.ctx.Bind(&req); err != nil {
			debugger.Error(debugger.ComponentPipeline, "failed to parse request body: %v", err)
			return errors.New().Skip(2).BadRequest().
				Summary("Failed to parse request body").
				Detail(err.Error()).
				Send(p.ctx)
		}
		newContent = req.Content
	}

	fileContent := &FileContent{
		Content: newContent,
	}

	if p.before != nil {
		debugger.PipelineStep("UpdateFileContent", "executing before hook")
		if err := p.before(p.ctx, fileContent); err != nil {
			debugger.Error(debugger.ComponentPipeline, "before hook failed: %v", err)
			return errors.New().Skip(2).InternalError().
				Summary("Update aborted by before hook").
				Detail(err.Error()).
				Send(p.ctx)
		}
	}

	debugger.PipelineStep("UpdateFileContent", "writing to file: %s", p.source)
	err := os.WriteFile(p.source, []byte(fileContent.Content), 0644)
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to write file: %v", err)
		return errors.New().Skip(2).InternalError().
			Summary("Failed to write file").
			Detail(err.Error()).
			Send(p.ctx)
	}

	fileInfo, err := os.Stat(p.source)
	if err != nil {
		debugger.Error(debugger.ComponentPipeline, "failed to stat file after write: %v", err)
	} else {
		fileContent.Size = fileInfo.Size()
	}

	if p.after != nil {
		debugger.PipelineStep("UpdateFileContent", "executing after hook")
		if err := p.after(p.ctx, fileContent); err != nil {
			debugger.Error(debugger.ComponentPipeline, "after hook failed: %v", err)
			return errors.New().Skip(2).InternalError().
				Summary("Update after hook failed").
				Detail(err.Error()).
				Send(p.ctx)
		}
	}

	if p.content != nil {
		*p.content = *fileContent
	}

	debugger.PipelineStep("UpdateFileContent", "returning JSON response")
	return p.ctx.JSON(200, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"content": fileContent.Content,
			"size":    fileContent.Size,
		},
		"message": "File content updated successfully",
	})
}

func (p *UpdateFileContentPipeline) respondError() error {
	debugger.Error(debugger.ComponentPipeline, "UpdateFileContent error: %v", p.err)
	if appErr, ok := p.err.(errors.Error); ok {
		return errors.New().Skip(3).
			Code(appErr.Details.Code).
			Summary("Update File Content Failed").
			Detail(p.err).
			Send(p.ctx)
	}
	return errors.New().Skip(3).BadRequest().Summary("Update File Content Failed").Detail(p.err).Send(p.ctx)
}

var _ = ReadFileContent
var _ = UpdateFileContent
