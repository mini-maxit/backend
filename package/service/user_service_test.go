package service_test

import (
	"testing"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
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
	cs := mock_service.NewMockContestService(ctrl)
	us := service.NewUserService(ur, cs)

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
	cs := mock_service.NewMockContestService(ctrl)
	us := service.NewUserService(ur, cs)

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
	cs := mock_service.NewMockContestService(ctrl)
	us := service.NewUserService(ur, cs)

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
		require.ErrorIs(t, err, errors.ErrForbidden)
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
	cs := mock_service.NewMockContestService(ctrl)
	us := service.NewUserService(ur, cs)
	paginationParams := schemas.PaginationParams{Limit: 10, Offset: 0, Sort: "id:asc"}

	t.Run("No users", func(t *testing.T) {
		ur.EXPECT().GetAll(
			tx,
			paginationParams.Limit,
			paginationParams.Offset,
			paginationParams.Sort,
		).Return([]models.User{}, int64(0), nil).Times(1)
		result, err := us.GetAll(tx, paginationParams)
		require.NoError(t, err)
		assert.Empty(t, result.Items)
		assert.Equal(t, 0, result.Pagination.TotalItems)
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
			paginationParams.Limit,
			paginationParams.Offset,
			paginationParams.Sort,
		).Return([]models.User{*user}, int64(1), nil).Times(1)
		result, err := us.GetAll(tx, paginationParams)
		require.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, user.Email, result.Items[0].Email)
		assert.Equal(t, user.Name, result.Items[0].Name)
		assert.Equal(t, user.Surname, result.Items[0].Surname)
		assert.Equal(t, user.Username, result.Items[0].Username)
		assert.Equal(t, user.ID, result.Items[0].ID)
		assert.Equal(t, user.Role, result.Items[0].Role)
		assert.Equal(t, 1, result.Pagination.TotalItems)
		assert.Equal(t, 1, result.Pagination.TotalItems)
	})
}

func TestChangeRole(t *testing.T) {
	tx := &gorm.DB{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	cs := mock_service.NewMockContestService(ctrl)
	us := service.NewUserService(ur, cs)

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
		require.ErrorIs(t, err, errors.ErrForbidden)
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
	cs := mock_service.NewMockContestService(ctrl)
	us := service.NewUserService(ur, cs)

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
		require.ErrorIs(t, err, errors.ErrForbidden)
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

func TestIsTaskAssignedToUser(t *testing.T) {
	tx := &gorm.DB{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	cs := mock_service.NewMockContestService(ctrl)
	us := service.NewUserService(ur, cs)

	userID := int64(1)
	taskID := int64(100)

	t.Run("Error getting task contests", func(t *testing.T) {
		cs.EXPECT().GetTaskContests(tx, taskID).Return(nil, assert.AnError).Times(1)
		isAssigned, err := us.IsTaskAssignedToUser(tx, userID, taskID)
		require.Error(t, err)
		assert.False(t, isAssigned)
	})

	t.Run("Task not assigned to any contest", func(t *testing.T) {
		cs.EXPECT().GetTaskContests(tx, taskID).Return([]int64{}, nil).Times(1)
		isAssigned, err := us.IsTaskAssignedToUser(tx, userID, taskID)
		require.NoError(t, err)
		assert.False(t, isAssigned)
	})

	t.Run("Task assigned to contests but user not participant", func(t *testing.T) {
		contestIDs := []int64{1, 2, 3}
		cs.EXPECT().GetTaskContests(tx, taskID).Return(contestIDs, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(1), userID).Return(false, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(2), userID).Return(false, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(3), userID).Return(false, nil).Times(1)
		isAssigned, err := us.IsTaskAssignedToUser(tx, userID, taskID)
		require.NoError(t, err)
		assert.False(t, isAssigned)
	})

	t.Run("Error checking user participation", func(t *testing.T) {
		contestIDs := []int64{1, 2}
		cs.EXPECT().GetTaskContests(tx, taskID).Return(contestIDs, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(1), userID).Return(false, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(2), userID).Return(false, assert.AnError).Times(1)
		isAssigned, err := us.IsTaskAssignedToUser(tx, userID, taskID)
		require.Error(t, err)
		assert.False(t, isAssigned)
	})

	t.Run("User is participant in first contest", func(t *testing.T) {
		contestIDs := []int64{1, 2, 3}
		cs.EXPECT().GetTaskContests(tx, taskID).Return(contestIDs, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(1), userID).Return(true, nil).Times(1)
		// Should return immediately after finding first match, no more calls
		isAssigned, err := us.IsTaskAssignedToUser(tx, userID, taskID)
		require.NoError(t, err)
		assert.True(t, isAssigned)
	})

	t.Run("User is participant in middle contest", func(t *testing.T) {
		contestIDs := []int64{1, 2, 3}
		cs.EXPECT().GetTaskContests(tx, taskID).Return(contestIDs, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(1), userID).Return(false, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(2), userID).Return(true, nil).Times(1)
		// Should return immediately after finding match, no call for contest 3
		isAssigned, err := us.IsTaskAssignedToUser(tx, userID, taskID)
		require.NoError(t, err)
		assert.True(t, isAssigned)
	})

	t.Run("User is participant in last contest", func(t *testing.T) {
		contestIDs := []int64{1, 2, 3}
		cs.EXPECT().GetTaskContests(tx, taskID).Return(contestIDs, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(1), userID).Return(false, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(2), userID).Return(false, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(3), userID).Return(true, nil).Times(1)
		isAssigned, err := us.IsTaskAssignedToUser(tx, userID, taskID)
		require.NoError(t, err)
		assert.True(t, isAssigned)
	})

	t.Run("Task in single contest and user is participant", func(t *testing.T) {
		contestIDs := []int64{5}
		cs.EXPECT().GetTaskContests(tx, taskID).Return(contestIDs, nil).Times(1)
		cs.EXPECT().IsUserParticipant(tx, int64(5), userID).Return(true, nil).Times(1)
		isAssigned, err := us.IsTaskAssignedToUser(tx, userID, taskID)
		require.NoError(t, err)
		assert.True(t, isAssigned)
	})
}
