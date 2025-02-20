package utils

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
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

// ApplyPaginationAndSort applies pagination and sort to the query.
//
// Values recived are guaranteed to be valid by middleware, so no error checking is needed.
func ApplyPaginationAndSort(tx *gorm.DB, limit, offset int, sortBy string) (*gorm.DB, error) {
	tx = tx.Limit(limit)
	tx = tx.Offset(offset)

	if sortBy != "" {
		sortFields := strings.Split(sortBy, ",")
		for _, sortField := range sortFields {
			sortFieldParts := strings.Split(sortField, ":")
			tx = tx.Order(sortFieldParts[0] + " " + sortFieldParts[1])
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
