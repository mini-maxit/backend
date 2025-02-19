package utils

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
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

func ApplyPaginationAndSort(tx *gorm.DB, limit, offset int, sortBy string) (*gorm.DB, error) {
	if limit >= 0 {
		tx = tx.Limit(limit)
	} else {
		err := httputils.QueryError{Filed: "limit", Detail: "limit must be grater or equal to 0"}
		return nil, err
	}
	if offset >= 0 {
		tx = tx.Offset(offset)
	} else {
		err := httputils.QueryError{Filed: "offset", Detail: "offset must be grater or equal to 0"}
		return nil, err
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
	return tx, nil
}

func usernameValidator(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	return re.MatchString(username)
}

func NewValidator() (*validator.Validate, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.RegisterValidation("username", usernameValidator)
	if err != nil {
		return nil, err
	}
	return validate, nil
}

func GetLimit(str string) (int, error) {
	limit, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, nil
	}
	return int(limit), nil
}

func GetOffset(str string) (int, error) {
	offset, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, nil
	}
	return int(offset), nil
}

func ValidateRoleAccess(current_role types.UserRole, acceptedRoles []types.UserRole) error {
	if !slices.Contains(acceptedRoles, current_role) {
		return errors.ErrNotAuthorized
	}
	return nil
}

func GetSort(str string) string {
	return ""
	// return str
}
