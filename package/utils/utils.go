package utils

import (
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/utils"
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

func ApplyQueryParams(tx *gorm.DB, queryParams map[string][]string) *gorm.DB {
	for key, values := range queryParams {
		switch key {
		case "limit":
			if len(values) > 0 {
				limit, err := strconv.Atoi(values[0])
				if err == nil {
					tx = tx.Limit(limit)
				} else {
					limit, _ := strconv.Atoi(utils.DefaultPaginationLimitStr)
					tx = tx.Limit(limit)
				}
			}
		case "offset":
			if len(values) > 0 {
				offset, err := strconv.Atoi(values[0])
				if err == nil {
					tx = tx.Offset(offset)
				} else {
					offset, _ := strconv.Atoi(utils.DefaultPaginationOffsetStr)
					tx = tx.Offset(offset)
				}
			}
		case "sort":
			if len(values) > 0 {
				tx = tx.Order(values[0])
			}
		default:
			if len(values) > 0 {
				tx = tx.Where(key, values)
			}
		}
	}

	return tx
}
