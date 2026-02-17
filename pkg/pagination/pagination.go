package pagination

import "math"

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

// PaginationQuery holds pagination input parsed from query strings.
type PaginationQuery struct {
	Page  int
	Limit int // 0 = return all items (no limit)
}

// PaginatedResult is the generic paginated response envelope.
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// FromQuery creates a PaginationQuery from raw page/limit values.
// Negative values are normalised. limit=0 means "return all".
func FromQuery(page, limit int) PaginationQuery {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	// limit == 0 is intentionally kept — means "all"
	return PaginationQuery{Page: page, Limit: limit}
}

// Offset returns the number of items to skip (for database queries).
// Returns 0 when Limit is 0 (all items).
func (q PaginationQuery) Offset() int64 {
	if q.Limit == 0 {
		return 0
	}
	return int64((q.Page - 1) * q.Limit)
}

// NewResult builds a PaginatedResult from the given items and total count.
func NewResult[T any](items []T, page, limit int, total int64) PaginatedResult[T] {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	} else {
		// limit == 0 → all items in one "page"
		if total > 0 {
			totalPages = 1
		}
	}

	return PaginatedResult[T]{
		Items:      items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
