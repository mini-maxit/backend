package service_test

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userServiceTest struct {
	tx          *gorm.DB
	config      *config.Config
	ur          repository.UserRepository
	userService service.UserService
	counter     int
}

func (ust *userServiceTest) createUser(t *testing.T, role types.UserRole) schemas.User {
	ust.counter++
	passHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userID, err := ust.ur.Create(ust.tx, &models.User{
		Name:         fmt.Sprintf("Test User %d", ust.counter),
		Surname:      fmt.Sprintf("Test Surname %d", ust.counter),
		Email:        fmt.Sprintf("email%d@email.com", ust.counter),
		Username:     fmt.Sprintf("testuser%d", ust.counter),
		Role:         role,
		PasswordHash: string(passHash),
	})
	require.NoError(t, err)

	userModel, err := ust.ur.Get(ust.tx, userID)
	require.NoError(t, err)

	user := schemas.User{
		ID:   userModel.ID,
		Role: userModel.Role,
	}
	return user
}

func newUserServiceTest() *userServiceTest {
	tx := &gorm.DB{}
	config := testutils.NewTestConfig()
	ur := testutils.NewMockUserRepository()
	us := service.NewUserService(ur)
	return &userServiceTest{
		tx:          tx,
		config:      config,
		ur:          ur,
		userService: us,
	}
}

func TestGetUserByEmail(t *testing.T) {
	ust := newUserServiceTest()

	t.Run("User does not exist", func(t *testing.T) {
		user, err := ust.userService.GetByEmail(ust.tx, "nonexistentemail")
		require.ErrorIs(t, err, errors.ErrUserNotFound)
		assert.Nil(t, user)
	})

	ust = newUserServiceTest()
	t.Run("User exists", func(t *testing.T) {
		user := &models.User{
			Name:         "Test User",
			Surname:      "Test Surname",
			Email:        "email@email.com",
			Username:     "testuser",
			PasswordHash: "password",
		}
		userID, err := ust.ur.Create(ust.tx, user)
		require.NoError(t, err)
		userResp, err := ust.userService.GetByEmail(ust.tx, user.Email)
		require.NoError(t, err)
		assert.NotNil(t, userResp)
		assert.Equal(t, userID, userResp.ID)
		assert.Equal(t, user.Email, userResp.Email)
		assert.Equal(t, user.Name, userResp.Name)
		assert.Equal(t, user.Surname, userResp.Surname)
		assert.Equal(t, user.Username, userResp.Username)
	})

	t.Run("Non existent email", func(t *testing.T) {
		user, err := ust.userService.GetByEmail(ust.tx, "nonexistentemail")
		require.ErrorIs(t, err, errors.ErrUserNotFound)
		assert.Nil(t, user)
	})
}

func TestGetUserByID(t *testing.T) {
	ust := newUserServiceTest()

	t.Run("User does not exist", func(t *testing.T) {
		user, err := ust.userService.Get(ust.tx, 0)
		require.ErrorIs(t, err, errors.ErrUserNotFound)
		assert.Nil(t, user)
	})
	ust = newUserServiceTest()
	t.Run("User exists", func(t *testing.T) {
		user := &models.User{
			Name:         "Test User",
			Surname:      "Test Surname",
			Email:        "email@email.com",
			Username:     "testuser",
			PasswordHash: "password",
		}
		userID, err := ust.ur.Create(ust.tx, user)
		require.NoError(t, err)
		userResp, err := ust.userService.Get(ust.tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, userResp)
		assert.Equal(t, userID, userResp.ID)
		assert.Equal(t, user.Email, userResp.Email)
		assert.Equal(t, user.Name, userResp.Name)
		assert.Equal(t, user.Surname, userResp.Surname)
		assert.Equal(t, user.Username, userResp.Username)
	})
}

func TestEditUser(t *testing.T) {
	ust := newUserServiceTest()
	adminUser := ust.createUser(t, types.UserRoleAdmin)
	studentUser := ust.createUser(t, types.UserRoleStudent)

	t.Run("User does not exist", func(t *testing.T) {
		err := ust.userService.Edit(ust.tx, adminUser, 0, &schemas.UserEdit{})
		require.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := ust.userService.Edit(ust.tx, studentUser, adminUser.ID, &schemas.UserEdit{})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not allowed", func(t *testing.T) {
		role := types.UserRoleAdmin
		err := ust.userService.Edit(ust.tx, studentUser, studentUser.ID, &schemas.UserEdit{Role: &role})
		require.ErrorIs(t, err, errors.ErrNotAllowed)
	})

	t.Run("Success", func(t *testing.T) {
		user := &models.User{
			Name:         "Test User",
			Surname:      "Test Surname",
			Email:        "email@email.com",
			Username:     "testuser",
			PasswordHash: "password",
		}
		userID, err := ust.ur.Create(ust.tx, user)
		require.NoError(t, err)
		newName := "New Name"
		updatedUser := &schemas.UserEdit{
			Name: &newName,
		}
		err = ust.userService.Edit(ust.tx, adminUser, userID, updatedUser)
		require.NoError(t, err)
		userResp, err := ust.userService.Get(ust.tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, userResp)
		assert.Equal(t, userID, userResp.ID)
		assert.Equal(t, newName, userResp.Name)
	})
}

func TestGetAllUsers(t *testing.T) {
	ust := newUserServiceTest()
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}

	t.Run("No users", func(t *testing.T) {
		users, err := ust.userService.GetAll(ust.tx, queryParams)
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("Users exist", func(t *testing.T) {
		user := &models.User{
			Name:         "Test User",
			Surname:      "Test Surname",
			Email:        "email@email.com",
			Username:     "testuser",
			PasswordHash: "password",
		}
		_, err := ust.ur.Create(ust.tx, user)
		require.NoError(t, err)
		users, err := ust.userService.GetAll(ust.tx, queryParams)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, user.Email, users[0].Email)
		assert.Equal(t, user.Name, users[0].Name)
		assert.Equal(t, user.Surname, users[0].Surname)
		assert.Equal(t, user.Username, users[0].Username)
		assert.Equal(t, user.ID, users[0].ID)
		assert.Equal(t, user.Role, users[0].Role)
	})
}

func TestChangeRole(t *testing.T) {
	ust := newUserServiceTest()
	adminUser := ust.createUser(t, types.UserRoleAdmin)
	studentUser := ust.createUser(t, types.UserRoleStudent)

	t.Run("User does not exist", func(t *testing.T) {
		err := ust.userService.ChangeRole(ust.tx, adminUser, 0, types.UserRoleAdmin)
		require.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := ust.userService.ChangeRole(ust.tx, studentUser, adminUser.ID, types.UserRoleAdmin)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Success", func(t *testing.T) {
		user := ust.createUser(t, types.UserRoleStudent)
		err := ust.userService.ChangeRole(ust.tx, adminUser, user.ID, types.UserRoleTeacher)
		require.NoError(t, err)
		userResp, err := ust.userService.Get(ust.tx, user.ID)
		require.NoError(t, err)
		if !assert.NotNil(t, userResp) {
			t.FailNow()
		}
		assert.Equal(t, types.UserRoleTeacher, userResp.Role)
	})
}

func TestChangePassword(t *testing.T) {
	ust := newUserServiceTest()
	user := ust.createUser(t, types.UserRoleStudent)
	adminUser := ust.createUser(t, types.UserRoleAdmin)
	randomUser := ust.createUser(t, types.UserRoleStudent)
	t.Run("User does not exist", func(t *testing.T) {
		err := ust.userService.ChangePassword(ust.tx, adminUser, 0, &schemas.UserChangePassword{})
		require.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := ust.userService.ChangePassword(ust.tx, randomUser, user.ID, &schemas.UserChangePassword{})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Invalid old password", func(t *testing.T) {
		err := ust.userService.ChangePassword(ust.tx, adminUser, user.ID, &schemas.UserChangePassword{
			OldPassword:        "invalidpassword",
			NewPassword:        "VeryStrongPass123!",
			NewPasswordConfirm: "VeryStrongPass123!"})
		require.ErrorIs(t, err, errors.ErrInvalidCredentials)
	})

	t.Run("Invalid data", func(t *testing.T) {
		err := ust.userService.ChangePassword(ust.tx, adminUser, user.ID, &schemas.UserChangePassword{
			OldPassword:        "password",
			NewPassword:        "VeryStrongPass123!",
			NewPasswordConfirm: "VeryStrongPass1234!"})
		assert.IsType(t, validator.ValidationErrors{}, err)
	})

	t.Run("Success", func(t *testing.T) {
		err := ust.userService.ChangePassword(ust.tx, adminUser, user.ID, &schemas.UserChangePassword{
			OldPassword:        "password",
			NewPassword:        "VeryStrongPass123!",
			NewPasswordConfirm: "VeryStrongPass123!"})
		require.NoError(t, err)
	})
}
