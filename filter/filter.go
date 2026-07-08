/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package filter

import (
	"strings"

	"github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/request"
)

type Order string

const (
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

type Filter struct {
	Keyword string
	Sort    string
	Order   Order
	Page    int
	Size    int
}

func New(ctx context.Context) *Builder {
	return &Builder{ctx: ctx}
}

type Builder struct {
	ctx context.Context
}

func (b *Builder) Build() Filter {
	keyword := request.Query(b.ctx, "keyword")
	sort := request.Query(b.ctx, "sort")
	order := Order(strings.ToLower(request.Query(b.ctx, "order")))
	if order != OrderAsc && order != OrderDesc {
		order = OrderAsc
	}
	page := request.QueryIntOrDefault(b.ctx, "page", 1)
	if page < 1 {
		page = 1
	}
	size := request.QueryIntOrDefault(b.ctx, "size", 20)
	if size < 1 {
		size = 20
	}
	return Filter{
		Keyword: keyword,
		Sort:    sort,
		Order:   order,
		Page:    page,
		Size:    size,
	}
}
