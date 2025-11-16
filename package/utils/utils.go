package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"gorm.io/gorm"
)

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

// UsernameValidator validates the username.
func UsernameValidator(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	return re.MatchString(username)
}

// PasswordValidator validates the password.
// Password must:
// * contain at least 8 characters
// * one uppercase letter
// * one lowercase letter
// * one digit
// * one special character.
func PasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	minPassLength := 8

	if len(password) < minPassLength {
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

// NewValidator creates a new validator with custom validators.
func NewValidator() (*validator.Validate, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Configure validator to use JSON tag names instead of struct field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := validate.RegisterValidation("username", UsernameValidator)
	if err != nil {
		return nil, err
	}
	err = validate.RegisterValidation("password", PasswordValidator)
	if err != nil {
		return nil, err
	}
	return validate, nil
}

// ValidateRoleAccess validates if the current role has access to the resource.
func ValidateRoleAccess(currentRole types.UserRole, acceptedRoles []types.UserRole) error {
	if !slices.Contains(acceptedRoles, currentRole) {
		return errors.ErrForbidden
	}
	return nil
}

// ValidateStruct validates that every field of a given struct is initialized
//
// Source: https://medium.com/@anajankow/fast-check-if-all-struct-fields-are-set-in-golang-bba1917213d2
func ValidateStruct(s any) error {
	// first make sure that the input is a struct
	// having any other type, especially a pointer to a struct,
	// might result in panic
	structType := reflect.TypeOf(s)
	if structType.Kind() != reflect.Struct {
		return errors.ErrExpectedStruct
	}

	// now go one by one through the fields and validate their value
	structVal := reflect.ValueOf(s)
	fieldNum := structVal.NumField()
	var err error

	for i := range fieldNum {
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
			err = fmt.Errorf("%w%s in not set; ", err, fieldName)
		}
	}

	return err
}
