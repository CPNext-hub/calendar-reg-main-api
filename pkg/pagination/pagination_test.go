package pagination

import "testing"

func TestFromQuery_Defaults(t *testing.T) {
	pq := FromQuery(0, 0)
	if pq.Page != DefaultPage {
		t.Errorf("expected page=%d, got %d", DefaultPage, pq.Page)
	}
	// limit=0 means "all", should stay 0
	if pq.Limit != 0 {
		t.Errorf("expected limit=0 (all), got %d", pq.Limit)
	}
}

func TestFromQuery_NegativePage(t *testing.T) {
	pq := FromQuery(-5, 20)
	if pq.Page != DefaultPage {
		t.Errorf("expected page=%d, got %d", DefaultPage, pq.Page)
	}
	if pq.Limit != 20 {
		t.Errorf("expected limit=20, got %d", pq.Limit)
	}
}

func TestFromQuery_NegativeLimit(t *testing.T) {
	pq := FromQuery(2, -10)
	if pq.Limit != DefaultLimit {
		t.Errorf("expected limit=%d, got %d", DefaultLimit, pq.Limit)
	}
}

func TestFromQuery_ExceedsMax(t *testing.T) {
	pq := FromQuery(1, 999)
	if pq.Limit != MaxLimit {
		t.Errorf("expected limit=%d, got %d", MaxLimit, pq.Limit)
	}
}

func TestFromQuery_NormalValues(t *testing.T) {
	pq := FromQuery(3, 25)
	if pq.Page != 3 {
		t.Errorf("expected page=3, got %d", pq.Page)
	}
	if pq.Limit != 25 {
		t.Errorf("expected limit=25, got %d", pq.Limit)
	}
}

func TestOffset_Normal(t *testing.T) {
	pq := PaginationQuery{Page: 3, Limit: 10}
	if got := pq.Offset(); got != 20 {
		t.Errorf("expected offset=20, got %d", got)
	}
}

func TestOffset_FirstPage(t *testing.T) {
	pq := PaginationQuery{Page: 1, Limit: 10}
	if got := pq.Offset(); got != 0 {
		t.Errorf("expected offset=0, got %d", got)
	}
}

func TestOffset_LimitZero(t *testing.T) {
	pq := PaginationQuery{Page: 5, Limit: 0}
	if got := pq.Offset(); got != 0 {
		t.Errorf("expected offset=0 for limit=0 (all), got %d", got)
	}
}

func TestNewResult_TotalPages(t *testing.T) {
	items := []string{"a", "b", "c"}
	r := NewResult(items, 1, 2, 5)
	if r.TotalPages != 3 {
		t.Errorf("expected totalPages=3, got %d", r.TotalPages)
	}
	if r.Total != 5 {
		t.Errorf("expected total=5, got %d", r.Total)
	}
	if len(r.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(r.Items))
	}
}

func TestNewResult_LimitZero(t *testing.T) {
	items := []string{"a", "b"}
	r := NewResult(items, 1, 0, 2)
	if r.TotalPages != 1 {
		t.Errorf("expected totalPages=1 for limit=0, got %d", r.TotalPages)
	}
}

func TestNewResult_EmptyItems(t *testing.T) {
	r := NewResult([]string{}, 1, 10, 0)
	if r.TotalPages != 0 {
		t.Errorf("expected totalPages=0 for empty, got %d", r.TotalPages)
	}
}

func TestNewResult_LimitZeroEmptyItems(t *testing.T) {
	r := NewResult([]string{}, 1, 0, 0)
	if r.TotalPages != 0 {
		t.Errorf("expected totalPages=0 for limit=0 and empty, got %d", r.TotalPages)
	}
}
