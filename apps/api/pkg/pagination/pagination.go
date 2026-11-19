package pagination

import (
	"errors"
	"math"

	"connectrpc.com/connect"

	"hasir/proto/gen/go/shared"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

func Normalize(req *shared.Pagination) (page, pageSize int) {
	page = DefaultPage
	pageSize = DefaultPageSize

	if req != nil {
		if req.GetPage() > 0 {
			page = int(req.GetPage())
		}
		if req.GetPageLimit() > 0 {
			pageSize = int(req.GetPageLimit())
		}
	}

	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	if page < 1 {
		page = DefaultPage
	}

	return page, pageSize
}

func Clamp(page, pageSize int) (normalizedPage, normalizedPageSize int) {
	normalizedPage = page
	normalizedPageSize = pageSize

	if normalizedPageSize < 1 {
		normalizedPageSize = DefaultPageSize
	}
	if normalizedPageSize > MaxPageSize {
		normalizedPageSize = MaxPageSize
	}
	if normalizedPage < 1 {
		normalizedPage = DefaultPage
	}

	return normalizedPage, normalizedPageSize
}

func ComputePageMeta(totalCount, page, pageSize int) (nextPage, totalPages int32, err error) {
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}

	pages := (totalCount + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	if pages > math.MaxInt32 {
		return 0, 0, connect.NewError(connect.CodeInternal, errors.New("total pages exceeds maximum value"))
	}

	if page < pages {
		if page+1 > math.MaxInt32 {
			return 0, 0, connect.NewError(connect.CodeInternal, errors.New("page number exceeds maximum value"))
		}
		nextPage = int32(page + 1) // #nosec G115 -- bounds checked above
	}

	return nextPage, int32(pages), nil // #nosec G115 -- bounds checked above
}
