package utils

import (
	"strconv"

	"gorm.io/gorm"
)

func TransactionPanicRecover(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	} else if tx != nil && tx.Error != nil {
		tx.Rollback()
	}
}

func ApplyFiltersAndSorting(tx *gorm.DB, filters map[string][]string, sort string) *gorm.DB {
	for key, values := range filters {
		switch key {
		case "limit":
			if len(values) > 0 {
				limit, err := strconv.Atoi(values[0])
				if err == nil {
					tx = tx.Limit(limit)
				}
			}
		case "offset":
			if len(values) > 0 {
				offset, err := strconv.Atoi(values[0])
				if err == nil {
					tx = tx.Offset(offset)
				}
			}
		default:
			if len(values) > 0 {
				tx = tx.Where(key, values)
			}
		}
	}
	if sort != "" {
		tx = tx.Order(sort)
	}
	return tx
}