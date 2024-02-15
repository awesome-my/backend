package awesomemy

import (
	"math"
	"net/http"
	"strconv"
)

// PageLimitOffsetFromRequest extracts the page, limit and offset from an HTTP request.
func PageLimitOffsetFromRequest(r *http.Request) (int, int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 20 {
		limit = 20
	}

	offset := 10 * (page - 1)

	return page, limit, offset
}

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	Count       int `json:"count"`
	Total       int `json:"total"`
}

func NewPaginationMeta(currentPage, count, total int) PaginationMeta {
	return PaginationMeta{
		CurrentPage: currentPage,
		TotalPages:  int(math.Ceil(float64(total) / float64(10))),
		Count:       count,
		Total:       total,
	}
}
