package service_test

import (
	"testing"
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)

	ur := mock_repository.NewMockUserRepository(ctrl)
	ss := mock_service.NewMockSessionService(ctrl)
	as := service.NewAuthService(ur, ss)

	userRegister := schemas.UserRegisterRequest{
		Name:            "name",
		Surname:         "surname",
		Email:           "email2@email.com",
		Username:        "username",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	t.Run("get user by email when user exists", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			userRegister.Email,
		).DoAndReturn(
			func(tx *gorm.DB, email string) (*models.User, error) {
				return &models.User{Email: email}, nil
			},
		).Times(1)
		response, err := as.Register(nil, userRegister)
		require.ErrorIs(t, err, myerrors.ErrUserAlreadyExists)
		assert.Nil(t, response)
	})

	t.Run("successful user registration", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			userRegister.Email,
		).Return(nil, nil).Times(1)

		ur.EXPECT().Create(
			gomock.Any(),
			gomock.Cond(func(user *models.User) bool {
				return user.Name == userRegister.Name &&
					user.Surname == userRegister.Surname &&
					user.Email == userRegister.Email &&
					user.Username == userRegister.Username &&
					user.PasswordHash != "" &&
					user.Role == types.UserRoleStudent
			}),
		).Return(int64(1), nil).Times(1)
		ss.EXPECT().Create(
			gomock.Any(),
			int64(1),
		).DoAndReturn(func(tx *gorm.DB, userID int64) (*schemas.Session, error) {
			return &schemas.Session{
				ID:        "session-id",
				UserID:    userID,
				UserRole:  string(types.UserRoleStudent),
				ExpiresAt: time.Now().Add(time.Hour * 24),
			}, nil
		},
		).Times(1)
		response, err := as.Register(nil, userRegister)
		require.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("failed to get user by email", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, gorm.ErrInvalidDB).Times(1)
		response, err := as.Register(nil, userRegister)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})

	t.Run("failed to create user", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			userRegister.Email,
		).Return(nil, nil).Times(1)

		ur.EXPECT().Create(
			gomock.Any(),
			gomock.Any(),
		).Return(int64(0), gorm.ErrInvalidDB).Times(1)

		response, err := as.Register(nil, userRegister)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})

	t.Run("failed to create session", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			userRegister.Email,
		).Return(nil, nil).Times(1)

		ur.EXPECT().Create(
			gomock.Any(),
			gomock.Any(),
		).Return(int64(1), nil).Times(1)

		ss.EXPECT().Create(
			gomock.Any(),
			int64(1),
		).Return(nil, gorm.ErrInvalidDB).Times(1)

		response, err := as.Register(nil, userRegister)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	ss := mock_service.NewMockSessionService(ctrl)
	as := service.NewAuthService(ur, ss)
	tx := &gorm.DB{}

	password := "Password123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &models.User{
		Name:         "name",
		Surname:      "surname",
		Email:        "email5@email.com",
		Username:     "username",
		PasswordHash: string(hash),
	}
	t.Run("get user by email when user does not exist", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			"nonexistent@email.com",
		).Return(nil, gorm.ErrRecordNotFound).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    "nonexistent@email.com",
			Password: "password",
		}

		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, myerrors.ErrUserNotFound)
		assert.Nil(t, response)
	})

	t.Run("compare password hash fails", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			user.Email,
		).Return(user, nil).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: "wrongpassword",
		}

		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, myerrors.ErrInvalidCredentials)
		assert.Nil(t, response)
	})

	t.Run("successful user login", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			user.Email,
		).Return(user, nil).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: password,
		}

		ss.EXPECT().Create(
			gomock.Any(),
			user.ID,
		).DoAndReturn(func(tx *gorm.DB, userID int64) (*schemas.Session, error) {
			return &schemas.Session{
				ID:        "session-id",
				UserID:    userID,
				UserRole:  string(types.UserRoleStudent),
				ExpiresAt: time.Now().Add(time.Hour * 24),
			}, nil
		}).Times(1)

		response, err := as.Login(tx, userLogin)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, user.ID, response.UserID)
	})

	t.Run("user repository fails unexpectedly", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, gorm.ErrInvalidTransaction).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    "unexpected@error.com",
			Password: "password",
		}

		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, gorm.ErrInvalidTransaction)
		assert.Nil(t, response)
	})

	t.Run("session service fails unexpectedly", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			user.Email,
		).Return(user, nil).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: password,
		}

		ss.EXPECT().Create(
			gomock.Any(),
			user.ID,
		).Return(nil, gorm.ErrInvalidDB).Times(1)

		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})
}
