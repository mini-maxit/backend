package utils_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		err := utils.ValidateStruct(d)
		require.Error(t, err)
	})

	t.Run("nil pointer", func(t *testing.T) {
		var d *Data
		err := utils.ValidateStruct(d)
		require.Error(t, err)
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

		err := utils.ValidateStruct(d)
		require.NoError(t, err)
	})

	t.Run("unititialized fields", func(t *testing.T) {
		var d struct {
			X Interfc
		}

		err := utils.ValidateStruct(d)
		require.Error(t, err)
	})
}

func TestValidateRoleAccess(t *testing.T) {
	t.Run("role is accepted", func(t *testing.T) {
		currentRole := types.UserRole("admin")
		acceptedRoles := []types.UserRole{"admin", "user"}

		err := utils.ValidateRoleAccess(currentRole, acceptedRoles)
		require.NoError(t, err)
	})

	t.Run("role is not accepted", func(t *testing.T) {
		currentRole := types.UserRole("guest")
		acceptedRoles := []types.UserRole{"admin", "user"}

		err := utils.ValidateRoleAccess(currentRole, acceptedRoles)
		require.Error(t, err)
		assert.Equal(t, errors.ErrNotAuthorized, err)
	})
}

func TestUsernameValidator(t *testing.T) {
	validate := validator.New()
	err := validate.RegisterValidation("username", utils.UsernameValidator)
	require.NoError(t, err)

	tests := []struct {
		username string
		valid    bool
	}{
		{"validUsername", true},
		{"validUsername123", true},
		{"1invalidUsername", false},
		{"invalid-username", false},
		{"", false},
	}

	for _, test := range tests {
		errs := validate.Var(test.username, "username")
		if test.valid {
			require.NoError(t, errs)
		} else {
			require.Error(t, errs)
		}
	}
}

func TestPasswordValidator(t *testing.T) {
	validate := validator.New()
	err := validate.RegisterValidation("password", utils.PasswordValidator)
	require.NoError(t, err)

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
				require.NoError(t, errs)
			} else {
				require.Error(t, errs)
			}
		})
	}
}
