package httputils

import "fmt"

type QueryError struct {
	Filed  string
	Detail string
}

func (e QueryError) Error() string {
	return fmt.Sprintf("Query error: %s: %s", e.Filed, e.Detail)
}

const MultipleQueryValues = "Multiple values for query parameter"

const DefaultPaginationLimitStr = "10"
const DefaultPaginationOffsetStr = "0"
const DefaultSortOrderField = "id:desc"
