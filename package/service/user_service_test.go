package service

import (
	"testing"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type userServiceTest struct {
	tx          *gorm.DB
	config      *config.Config
	ur          repository.UserRepository
	userService UserService
}

func newUserServiceTest() *userServiceTest {
	tx := &gorm.DB{}
	config := testutils.NewTestConfig()
	ur := testutils.NewMockUserRepository()
	us := NewUserService(ur)
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
		user, err := ust.userService.GetUserByEmail(ust.tx, "nonexistentemail")
		assert.ErrorIs(t, err, ErrUserNotFound)
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
		userId, err := ust.ur.CreateUser(ust.tx, user)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		userResp, err := ust.userService.GetUserByEmail(ust.tx, user.Email)
		assert.NoError(t, err)
		if !assert.NotNil(t, userResp) {
			t.FailNow()
		}
		assert.Equal(t, userId, userResp.Id)
		assert.Equal(t, user.Email, userResp.Email)
		assert.Equal(t, user.Name, userResp.Name)
		assert.Equal(t, user.Surname, userResp.Surname)
		assert.Equal(t, user.Username, userResp.Username)
	})

}

func TestGetUserById(t *testing.T) {
	ust := newUserServiceTest()

	t.Run("User does not exist", func(t *testing.T) {
		user, err := ust.userService.GetUserById(ust.tx, 0)
		assert.ErrorIs(t, err, ErrUserNotFound)
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
		userId, err := ust.ur.CreateUser(ust.tx, user)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		userResp, err := ust.userService.GetUserById(ust.tx, userId)
		assert.NoError(t, err)
		if !assert.NotNil(t, userResp) {
			t.FailNow()
		}
		assert.Equal(t, userId, userResp.Id)
		assert.Equal(t, user.Email, userResp.Email)
		assert.Equal(t, user.Name, userResp.Name)
		assert.Equal(t, user.Surname, userResp.Surname)
		assert.Equal(t, user.Username, userResp.Username)
	})

}

func TestEditUser(t *testing.T) {
	ust := newUserServiceTest()

	t.Run("User does not exist", func(t *testing.T) {
		err := ust.userService.EditUser(ust.tx, 0, &schemas.UserEdit{})
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("Success", func(t *testing.T) {
		user := &models.User{
			Name:         "Test User",
			Surname:      "Test Surname",
			Email:        "email@email.com",
			Username:     "testuser",
			PasswordHash: "password",
		}
		userId, err := ust.ur.CreateUser(ust.tx, user)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		newName := "New Name"
		updatedUser := &schemas.UserEdit{
			Name: &newName,
		}
		err = ust.userService.EditUser(ust.tx, userId, updatedUser)
		assert.NoError(t, err)
		userResp, err := ust.userService.GetUserById(ust.tx, userId)
		assert.NoError(t, err)
		if !assert.NotNil(t, userResp) {
			t.FailNow()
		}
		assert.Equal(t, userId, userResp.Id)
		assert.Equal(t, newName, userResp.Name)
	})
}

func TestGetAllUsers(t *testing.T) {
	ust := newUserServiceTest()
	queryParams := map[string]string{"limit": "10", "offset": "0", "sort": "id:asc"}

	t.Run("No users", func(t *testing.T) {
		users, err := ust.userService.GetAllUsers(ust.tx, queryParams)
		assert.NoError(t, err)
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
		_, err := ust.ur.CreateUser(ust.tx, user)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		users, err := ust.userService.GetAllUsers(ust.tx, queryParams)
		assert.NoError(t, err)
		if !assert.Len(t, users, 1) {
			t.FailNow()
		}
		assert.Equal(t, user.Email, users[0].Email)
		assert.Equal(t, user.Name, users[0].Name)
		assert.Equal(t, user.Surname, users[0].Surname)
		assert.Equal(t, user.Username, users[0].Username)
		assert.Equal(t, user.Id, users[0].Id)
		assert.Equal(t, string(user.Role), users[0].Role)
	})
}
