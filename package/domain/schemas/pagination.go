package schemas

type PaginationParams struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=100"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"omitempty,sort"`
}

// PaginationMetadata contains metadata about paginated results
type PaginationMetadata struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalItems  int `json:"totalItems"`
	TotalPages  int `json:"totalPages"`
}

// PaginatedResult wraps paginated data with metadata
type PaginatedResult[T any] struct {
	Pagination PaginationMetadata `json:"pagination"`
	Items      T                  `json:"items"`
}

// NewPaginationMetadata creates pagination metadata from offset, limit, and total count
func NewPaginationMetadata(offset, limit, totalItems int) PaginationMetadata {
	// Defensive defaults
	if limit <= 0 {
		// No paging (or invalid page size) — treat as single page
		return PaginationMetadata{
			CurrentPage: 1,
			PageSize:    limit,
			TotalItems:  totalItems,
			TotalPages:  0,
		}
	}

	if offset < 0 {
		offset = 0
	}

	// totalPages first (ceiling division)
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + limit - 1) / limit
	}

	// derive current page from offset
	currentPage := (offset / limit) + 1

	// if there are no items, ensure currentPage stays 1
	if totalPages == 0 {
		currentPage = 1
	} else if currentPage > totalPages {
		// clamp to the last page if offset points past the end
		currentPage = totalPages
	}

	return PaginationMetadata{
		CurrentPage: currentPage,
		PageSize:    limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
	}
}

// NewPaginatedResult creates a paginated response
func NewPaginatedResult[T any](data T, offset, limit int, totalItems int64) PaginatedResult[T] {
	return PaginatedResult[T]{
		Pagination: NewPaginationMetadata(offset, limit, int(totalItems)),
		Items:      data,
	}
}
