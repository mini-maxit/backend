package utils

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/stretchr/testify/assert"
)

type Data struct {
	Y            int
	ptr          *string
	structMember Impl
	ifMember     Interfc
	mapMember    map[string]any
	sliceMember  []string
}

type Interfc interface {
	DoIt()
}

type Impl struct {
	implField map[string]string
}

func (i Impl) DoIt() {}

func TestValidateStructEmpty(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		var d Data
		err := ValidateStruct(d)
		if !assert.Error(t, err) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var d *Data
		err := ValidateStruct(d)
		if !assert.Error(t, err) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("different field types", func(t *testing.T) {

		emptyString := ""

		d := Data{
			Y:   1,
			ptr: &emptyString,

			structMember: Impl{implField: make(map[string]string)},

			ifMember: Impl{},

			mapMember:   make(map[string]any),
			sliceMember: make([]string, 0),
		}

		err := ValidateStruct(d)
		if !assert.NoError(t, err) {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("unititialized fields", func(t *testing.T) {
		var d struct {
			X Interfc
		}

		err := ValidateStruct(d)
		if !assert.Error(t, err) {
			t.Fatalf("expected error, got nil")
		}
	})

}
func TestValidateRoleAccess(t *testing.T) {
	t.Run("role is accepted", func(t *testing.T) {
		currentRole := types.UserRole("admin")
		acceptedRoles := []types.UserRole{"admin", "user"}

		err := ValidateRoleAccess(currentRole, acceptedRoles)
		assert.NoError(t, err)
	})

	t.Run("role is not accepted", func(t *testing.T) {
		currentRole := types.UserRole("guest")
		acceptedRoles := []types.UserRole{"admin", "user"}

		err := ValidateRoleAccess(currentRole, acceptedRoles)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrNotAuthorized, err)
	})
}
func TestGetLimit(t *testing.T) {
	t.Run("valid limit", func(t *testing.T) {
		limitStr := "10"
		expectedLimit := 10

		limit, err := GetLimit(limitStr)
		assert.NoError(t, err)
		assert.Equal(t, expectedLimit, limit)
	})

	t.Run("invalid limit", func(t *testing.T) {
		limitStr := "invalid"

		limit, err := GetLimit(limitStr)
		assert.Error(t, err)
		assert.Equal(t, 0, limit)
	})
}

func TestGetOffset(t *testing.T) {
	t.Run("valid offset", func(t *testing.T) {
		offsetStr := "20"
		expectedOffset := 20

		offset, err := GetOffset(offsetStr)
		assert.NoError(t, err)
		assert.Equal(t, expectedOffset, offset)
	})

	t.Run("invalid offset", func(t *testing.T) {
		offsetStr := "invalid"

		offset, err := GetOffset(offsetStr)
		assert.Error(t, err)
		assert.Equal(t, 0, offset)
	})
}
func TestUsernameValidator(t *testing.T) {
	validate := validator.New()
	err := validate.RegisterValidation("username", usernameValidator)
	assert.NoError(t, err)

	tests := []struct {
		username string
		valid    bool
	}{
		{"validUsername", true},
		{"valid_username123", true},
		{"1invalidUsername", false},
		{"invalid-username", false},
		{"", false},
	}

	for _, test := range tests {
		errs := validate.Var(test.username, "username")
		if test.valid {
			assert.NoError(t, errs)
		} else {
			assert.Error(t, errs)
		}
	}
}

func TestPasswordValidator(t *testing.T) {
	validate := validator.New()
	err := validate.RegisterValidation("password", passwordValidator)
	assert.NoError(t, err)

	tests := []struct {
		password string
		valid    bool
	}{
		{"Valid1Password!", true},
		{"sht1A!", false},
		{"nouppercase1!", false},
		{"NOLOWERCASE1!", false},
		{"NoSpecialChar1", false},
		{"NoDigit!", false},
	}

	for _, test := range tests {
		t.Run(test.password, func(t *testing.T) {

			errs := validate.Var(test.password, "password")
			if test.valid {
				assert.NoError(t, errs)
			} else {
				assert.Error(t, errs)
			}
		})
	}
}
