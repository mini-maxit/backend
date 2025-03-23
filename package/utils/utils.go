package utils

import (
	"fmt"
	"reflect"
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

func passwordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	upperCase := regexp.MustCompile(`[A-Z]`)
	lowerCase := regexp.MustCompile(`[a-z]`)
	digit := regexp.MustCompile(`[0-9]`)
	specialChar := regexp.MustCompile(`[#!?@$%^&*-]`)

	return upperCase.MatchString(password) &&
		lowerCase.MatchString(password) &&
		digit.MatchString(password) &&
		specialChar.MatchString(password)
}

func NewValidator() (*validator.Validate, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.RegisterValidation("username", usernameValidator)
	if err != nil {
		return nil, err
	}
	err = validate.RegisterValidation("password", passwordValidator)
	if err != nil {
		return nil, err
	}
	return validate, nil
}

func GetLimit(str string) (int, error) {
	limit, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(limit), nil
}

func GetOffset(str string) (int, error) {
	offset, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
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

// ValidateStruct validates that every field of a given struct is initialized
//
// Source: https://medium.com/@anajankow/fast-check-if-all-struct-fields-are-set-in-golang-bba1917213d2
func ValidateStruct(s any) (err error) {
	// first make sure that the input is a struct
	// having any other type, especially a pointer to a struct,
	// might result in panic
	structType := reflect.TypeOf(s)
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf("input param should be a struct")
	}

	// now go one by one through the fields and validate their value
	structVal := reflect.ValueOf(s)
	fieldNum := structVal.NumField()

	for i := 0; i < fieldNum; i++ {
		// Field(i) returns i'th value of the struct
		field := structVal.Field(i)
		fieldName := structType.Field(i).Name

		// CAREFUL! IsZero interprets empty strings and int equal 0 as a zero value.
		// To check only if the pointers have been initialized,
		// you can check the kind of the field:
		// if field.Kind() == reflect.Pointer { // check }

		// IsZero panics if the value is invalid.
		// Most functions and methods never return an invalid Value.
		isSet := field.IsValid() && !field.IsZero()

		if !isSet {
			err = fmt.Errorf("%v%s in not set; ", err, fieldName)
		}

	}

	return err
}
