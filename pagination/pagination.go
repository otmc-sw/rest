/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package pagination

import (
	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/request"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Page struct {
	Number int
	Size   int
}

func (p Page) Offset() int {
	return (p.Number - 1) * p.Size
}

func (p Page) Limit() int {
	return p.Size
}

func New(ctx context.Context) *Builder {
	return &Builder{ctx: ctx}
}

type Builder struct {
	ctx         context.Context
	defaultSize int
}

func (b *Builder) DefaultSize(size int) *Builder {
	b.defaultSize = size
	return b
}

func (b *Builder) Page() Page {
	size := b.defaultSize
	if size <= 0 {
		size = DefaultPageSize
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	number := request.QueryIntOrDefault(b.ctx, "page", DefaultPage)
	if number < 1 {
		number = DefaultPage
	}
	reqSize := request.QueryIntOrDefault(b.ctx, "size", size)
	if reqSize < 1 {
		reqSize = size
	}
	if reqSize > MaxPageSize {
		reqSize = MaxPageSize
	}
	return Page{Number: number, Size: reqSize}
}

type Meta struct {
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func NewMeta(p Page, total int64) Meta {
	totalPages := int64(0)
	if p.Size > 0 {
		totalPages = (total + int64(p.Size) - 1) / int64(p.Size)
	}
	return Meta{
		Page:       p.Number,
		Size:       p.Size,
		Total:      total,
		TotalPages: totalPages,
	}
}
