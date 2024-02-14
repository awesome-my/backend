package awesomemy

import (
	"math"
	"net/http"
	"strconv"
)

// PageAndOffsetFromRequest extracts the page and offset from an HTTP request.
func PageAndOffsetFromRequest(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	offset := 10 * (page - 1)

	return page, offset
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
		TotalPages:  int(math.Ceil(float64(count) / float64(10))),
		Count:       count,
		Total:       total,
	}
}
