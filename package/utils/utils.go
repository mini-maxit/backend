package utils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
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

func ApplyPaginationAndSort(tx *gorm.DB, limitStr, offsetStr, sortBy string) *gorm.DB {
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit > 0 {
		tx = tx.Limit(limit)
	}
	if offset > 0 {
		tx = tx.Offset(offset)
	}
	if sortBy != "" {
		sortFields := strings.Split(sortBy, ",")
		for _, sortField := range sortFields {
			sortFieldParts := strings.Split(sortField, ":")
			if len(sortFieldParts) == 2 {
				tx = tx.Order(sortFieldParts[0] + " " + sortFieldParts[1])
			} else {
				tx = tx.Order(sortFieldParts[0] + " " + httputils.DefaultSortOrder)
			}	
		}
	}
	return tx
}


func usernameValidator(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	return re.MatchString(username)
}

func NewValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("username", usernameValidator)
	return validate
}
