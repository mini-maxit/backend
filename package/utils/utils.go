package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
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
