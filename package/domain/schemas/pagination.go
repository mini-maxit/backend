package schemas

type PaginationParams struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=100"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"omitempty,sort"`
}

// PaginationMetadata contains metadata about paginated results
type PaginationMetadata struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
}

// PaginatedResponse wraps paginated data with metadata
type PaginatedResponse[T any] struct {
	Pagination PaginationMetadata `json:"pagination"`
	Items      T                  `json:"items"`
}

// NewPaginationMetadata creates pagination metadata from offset, limit, and total count
func NewPaginationMetadata(offset, limit, totalItems int) PaginationMetadata {
	currentPage := 1
	if limit > 0 {
		currentPage = (offset / limit) + 1
	}

	totalPages := 0
	if limit > 0 && totalItems > 0 {
		totalPages = (totalItems + limit - 1) / limit // Ceiling division
	}

	return PaginationMetadata{
		CurrentPage: currentPage,
		PageSize:    limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
	}
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse[T any](data T, offset, limit, totalItems int) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Pagination: NewPaginationMetadata(offset, limit, totalItems),
		Items:      data,
	}
}
