package schemas_test

import (
	"testing"

	"github.com/mini-maxit/backend/package/domain/schemas"
)

func TestNewPaginationMetadata(t *testing.T) {
	tests := []struct {
		name       string
		offset     int
		limit      int
		totalItems int
		want       schemas.PaginationMetadata
	}{
		{
			name:       "First page with 10 items per page, 25 total",
			offset:     0,
			limit:      10,
			totalItems: 25,
			want: schemas.PaginationMetadata{
				CurrentPage: 1,
				PageSize:    10,
				TotalItems:  25,
				TotalPages:  3,
			},
		},
		{
			name:       "Second page with 10 items per page, 25 total",
			offset:     10,
			limit:      10,
			totalItems: 25,
			want: schemas.PaginationMetadata{
				CurrentPage: 2,
				PageSize:    10,
				TotalItems:  25,
				TotalPages:  3,
			},
		},
		{
			name:       "Third page with 10 items per page, 25 total",
			offset:     20,
			limit:      10,
			totalItems: 25,
			want: schemas.PaginationMetadata{
				CurrentPage: 3,
				PageSize:    10,
				TotalItems:  25,
				TotalPages:  3,
			},
		},
		{
			name:       "Exact page boundary - 30 items with 10 per page",
			offset:     0,
			limit:      10,
			totalItems: 30,
			want: schemas.PaginationMetadata{
				CurrentPage: 1,
				PageSize:    10,
				TotalItems:  30,
				TotalPages:  3,
			},
		},
		{
			name:       "No items",
			offset:     0,
			limit:      10,
			totalItems: 0,
			want: schemas.PaginationMetadata{
				CurrentPage: 1,
				PageSize:    10,
				TotalItems:  0,
				TotalPages:  0,
			},
		},
		{
			name:       "Large offset with 5 items per page, 100 total",
			offset:     50,
			limit:      5,
			totalItems: 100,
			want: schemas.PaginationMetadata{
				CurrentPage: 11,
				PageSize:    5,
				TotalItems:  100,
				TotalPages:  20,
			},
		},
		{
			name:       "Single item per page",
			offset:     5,
			limit:      1,
			totalItems: 10,
			want: schemas.PaginationMetadata{
				CurrentPage: 6,
				PageSize:    1,
				TotalItems:  10,
				TotalPages:  10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schemas.NewPaginationMetadata(tt.offset, tt.limit, tt.totalItems)
			if got != tt.want {
				t.Errorf("NewPaginationMetadata() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	type testData struct {
		Name string
		ID   int
	}

	data := []testData{
		{Name: "Item1", ID: 1},
		{Name: "Item2", ID: 2},
	}

	response := schemas.NewPaginatedResponse(data, 10, 5, 25)

	if response.Pagination.CurrentPage != 3 {
		t.Errorf("Expected current page 3, got %d", response.Pagination.CurrentPage)
	}

	if response.Pagination.PageSize != 5 {
		t.Errorf("Expected page size 5, got %d", response.Pagination.PageSize)
	}

	if response.Pagination.TotalItems != 25 {
		t.Errorf("Expected total items 25, got %d", response.Pagination.TotalItems)
	}

	if response.Pagination.TotalPages != 5 {
		t.Errorf("Expected total pages 5, got %d", response.Pagination.TotalPages)
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 data items, got %d", len(response.Data))
	}

	if response.Data[0].Name != "Item1" {
		t.Errorf("Expected first item name 'Item1', got '%s'", response.Data[0].Name)
	}
}
