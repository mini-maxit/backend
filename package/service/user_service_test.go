package service_test

import (
	"testing"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestGetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	tx := &gorm.DB{}
	ur := mock_repository.NewMockUserRepository(ctrl)
	us := service.NewUserService(ur)

	t.Run("User does not exist", func(t *testing.T) {
		ur.EXPECT().GetByEmail(tx, "nonexistentemail").Return(nil, gorm.ErrRecordNotFound).Times(1)
		user, err := us.GetByEmail(tx, "nonexistentemail")
		require.ErrorIs(t, err, errors.ErrUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("User exists", func(t *testing.T) {
		user := &models.User{
			ID:           int64(1),
			Name:         "Test User",
			Surname:      "Test Surname",
			Email:        "email@email.com",
			Username:     "testuser",
			PasswordHash: "password",
		}
		ur.EXPECT().GetByEmail(tx, user.Email).Return(user, nil).Times(1)
		userResp, err := us.GetByEmail(tx, user.Email)
		require.NoError(t, err)
		assert.NotNil(t, userResp)
		assert.Equal(t, user.ID, userResp.ID)
		assert.Equal(t, user.Email, userResp.Email)
		assert.Equal(t, user.Name, userResp.Name)
		assert.Equal(t, user.Surname, userResp.Surname)
		assert.Equal(t, user.Username, userResp.Username)
	})
}

func TestGetUserByID(t *testing.T) {
	tx := &gorm.DB{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	us := service.NewUserService(ur)

	t.Run("User does not exist", func(t *testing.T) {
		ur.EXPECT().Get(tx, int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)
		user, err := us.Get(tx, int64(1))
		require.ErrorIs(t, err, errors.ErrUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("User exists", func(t *testing.T) {
		user := &models.User{
			ID:           int64(1),
			Name:         "Test User",
			Surname:      "Test Surname",
			Email:        "email@email.com",
			Username:     "testuser",
			PasswordHash: "password",
		}
		ur.EXPECT().Get(tx, user.ID).Return(user, nil).Times(1)
		userResp, err := us.Get(tx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, userResp)
		assert.Equal(t, user.ID, userResp.ID)
		assert.Equal(t, user.Email, userResp.Email)
		assert.Equal(t, user.Name, userResp.Name)
		assert.Equal(t, user.Surname, userResp.Surname)
		assert.Equal(t, user.Username, userResp.Username)
	})
}

func TestEditUser(t *testing.T) {
	tx := &gorm.DB{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	us := service.NewUserService(ur)

	adminUser := schemas.User{
		ID:   1,
		Role: types.UserRoleAdmin,
	}
	studentUser := schemas.User{
		ID:   2,
		Role: types.UserRoleStudent,
	}

	t.Run("User does not exist", func(t *testing.T) {
		ur.EXPECT().Get(tx, studentUser.ID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := us.Edit(tx, adminUser, studentUser.ID, &schemas.UserEdit{})
		require.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := us.Edit(tx, studentUser, adminUser.ID, &schemas.UserEdit{})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not allowed", func(t *testing.T) {
		role := types.UserRoleAdmin
		ur.EXPECT().Get(tx, studentUser.ID).Return(&models.User{
			ID:   studentUser.ID,
			Role: types.UserRoleStudent,
		}, nil).Times(1)
		err := us.Edit(
			tx,
			studentUser,
			studentUser.ID,
			&schemas.UserEdit{Role: &role})
		require.ErrorIs(t, err, errors.ErrNotAllowed)
	})

	t.Run("Success", func(t *testing.T) {
		user := &models.User{
			ID:           3,
			Name:         "Test User",
			Surname:      "Test Surname",
			Email:        "email@email.com",
			Username:     "testuser",
			PasswordHash: "password",
		}
		newName := "New Name"
		updatedUser := &schemas.UserEdit{
			Name: &newName,
		}
		ur.EXPECT().Get(tx, user.ID).Return(user, nil).Times(1)
		ur.EXPECT().Edit(tx, user).Return(nil).Times(1)
		err := us.Edit(tx, adminUser, user.ID, updatedUser)
		require.NoError(t, err)
	})
}

func TestGetAllUsers(t *testing.T) {
	tx := &gorm.DB{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	us := service.NewUserService(ur)
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}

	t.Run("No users", func(t *testing.T) {
		ur.EXPECT().GetAll(
			tx,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.User{}, nil).Times(1)
		users, err := us.GetAll(tx, queryParams)
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
		ur.EXPECT().GetAll(
			tx,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.User{*user}, nil).Times(1)
		users, err := us.GetAll(tx, queryParams)
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
	tx := &gorm.DB{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	us := service.NewUserService(ur)

	adminUser := schemas.User{
		ID:   1,
		Role: types.UserRoleAdmin,
	}
	studentUser := schemas.User{
		ID:   2,
		Role: types.UserRoleStudent,
	}

	t.Run("User does not exist", func(t *testing.T) {
		ur.EXPECT().Get(tx, int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := us.ChangeRole(tx, adminUser, int64(1), types.UserRoleAdmin)
		require.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := us.ChangeRole(tx, studentUser, adminUser.ID, types.UserRoleAdmin)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Success", func(t *testing.T) {
		ur.EXPECT().Get(tx, studentUser.ID).Return(&models.User{
			ID:   studentUser.ID,
			Role: types.UserRoleStudent,
		}, nil).Times(1)
		ur.EXPECT().Edit(tx, &models.User{
			ID:   studentUser.ID,
			Role: types.UserRoleTeacher,
		}).Times(1)
		err := us.ChangeRole(tx, adminUser, studentUser.ID, types.UserRoleTeacher)
		require.NoError(t, err)
	})
}

func TestChangePassword(t *testing.T) {
	tx := &gorm.DB{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	us := service.NewUserService(ur)

	password := "password"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &models.User{
		ID:           1,
		Role:         types.UserRoleStudent,
		PasswordHash: string(hash),
	}
	newPassword := "VeryStrongPass123!"
	adminUser := schemas.User{
		ID:   2,
		Role: types.UserRoleAdmin,
	}
	randomUser := schemas.User{
		ID:   3,
		Role: types.UserRoleStudent,
	}
	t.Run("User does not exist", func(t *testing.T) {
		ur.EXPECT().Get(tx, int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := us.ChangePassword(tx, adminUser, 1, &schemas.UserChangePassword{})
		require.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := us.ChangePassword(tx, randomUser, user.ID, &schemas.UserChangePassword{})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Invalid old password", func(t *testing.T) {
		ur.EXPECT().Get(tx, user.ID).Return(user, nil).Times(1)
		err := us.ChangePassword(tx, adminUser, user.ID, &schemas.UserChangePassword{
			OldPassword:        "invalidpassword",
			NewPassword:        newPassword,
			NewPasswordConfirm: newPassword})
		require.ErrorIs(t, err, errors.ErrInvalidCredentials)
	})

	t.Run("Invalid data", func(t *testing.T) {
		ur.EXPECT().Get(tx, user.ID).Return(user, nil).Times(1)
		err := us.ChangePassword(tx, adminUser, user.ID, &schemas.UserChangePassword{
			OldPassword:        password,
			NewPassword:        newPassword,
			NewPasswordConfirm: newPassword + "123"})
		require.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		ur.EXPECT().Get(tx, user.ID).Return(user, nil).Times(1)
		ur.EXPECT().Edit(tx, gomock.Any()).Return(nil).Times(1)
		err := us.ChangePassword(tx, adminUser, user.ID, &schemas.UserChangePassword{
			OldPassword:        password,
			NewPassword:        newPassword,
			NewPasswordConfirm: newPassword})
		require.NoError(t, err)
	})
}
