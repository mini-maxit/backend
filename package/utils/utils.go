package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/package/domain/models"
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

func UserRoleToString(userRole models.UserRole) string {
	switch userRole {
	case 0:
		return "guest"
	case 1:
		return "student"
	case 2:
		return "teacher"
	case 3:
		return "admin"
	default:
		return "unknown"
	}
}
