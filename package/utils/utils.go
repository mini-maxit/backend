package utils

import (
	// "fmt"
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

func ApplyFilters(tx *gorm.DB, filters map[string][]string) *gorm.DB {
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
	return tx
}

