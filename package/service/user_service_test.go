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
	savePoint   string
}

func newUserServiceTest(t *testing.T) *userServiceTest {
	tx := testutils.NewTestTx(t)
	config := testutils.NewTestConfig()
	ur, err := repository.NewUserRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	us := NewUserService(ur)
	savePoint := "savepoint"
	tx.SavePoint(savePoint)
	return &userServiceTest{
		tx:          tx,
		config:      config,
		ur:          ur,
		userService: us,
		savePoint:   savePoint,
	}
}

func (ust *userServiceTest) RollbackToSavepoint() {
	ust.tx.RollbackTo(ust.savePoint)
}

func TestGetUserByEmail(t *testing.T) {
	ust := newUserServiceTest(t)
	defer ust.tx.Rollback()

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
		ust.RollbackToSavepoint()
	})

	t.Run("User does not exist", func(t *testing.T) {
		user, err := ust.userService.GetUserByEmail(ust.tx, "nonexistentemail")
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
	})
}

func TestGetUserById(t *testing.T) {
	ust := newUserServiceTest(t)
	defer ust.tx.Rollback()

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
		ust.RollbackToSavepoint()
	})

	t.Run("User does not exist", func(t *testing.T) {
		user, err := ust.userService.GetUserById(ust.tx, 0)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
	})
}

func TestEditUser(t *testing.T) {
	ust := newUserServiceTest(t)
	defer ust.tx.Rollback()

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
		ust.RollbackToSavepoint()
	})

	t.Run("User does not exist", func(t *testing.T) {
		err := ust.userService.EditUser(ust.tx, 0, &schemas.UserEdit{})
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}
