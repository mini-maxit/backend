// TODO: Add missing test cases, for edge cases, and error cases
package service_test

import (
	"github.com/mini-maxit/backend/internal/database"
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
	"gorm.io/gorm"
)

func TestCreateGroup(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	t.Run("Success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		gr.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(1), nil).Times(1)
		groupID, err := gs.Create(tx, *currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		groupID, err := gs.Create(tx, *currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Equal(t, int64(0), groupID)
	})
}

func TestDeleteGroup(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	t.Run("Success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		group := &models.Group{
			ID:        int64(1),
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		}
		gr.EXPECT().Delete(gomock.Any(), group.ID).Return(nil).Times(1)
		gr.EXPECT().Get(gomock.Any(), group.ID).Return(group, nil).Times(1)
		err := gs.Delete(tx, *currentUser, group.ID)
		require.NoError(t, err)
	})

	t.Run("Not authorized student", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		err := gs.Delete(tx, *currentUser, 2)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		currentUser := &schemas.User{ID: 3, Role: types.UserRoleTeacher}
		gr.EXPECT().Get(gomock.Any(), int64(2)).Return(&models.Group{
			ID:        int64(2),
			Name:      "Test Group",
			CreatedBy: 1,
		}, nil).Times(1)
		err := gs.Delete(tx, *currentUser, 2)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetAllGroup(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}
	t.Run("No groups", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		gr.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.Group{}, nil).Times(1)
		groups, err := gs.GetAll(tx, *currentUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, groups)
	})

	t.Run("Success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		gr.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.Group{
			{
				ID:        1,
				Name:      "Test Group",
				CreatedBy: currentUser.ID,
			},
		}, nil).Times(1)
		groups, err := gs.GetAll(tx, *currentUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, groups)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		groups, err := gs.GetAll(tx, *currentUser, queryParams)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, groups)
	})

	t.Run("Success for teacher", func(t *testing.T) {
		currentUser := &schemas.User{ID: 3, Role: types.UserRoleTeacher}
		gr.EXPECT().GetAllForTeacher(
			gomock.Any(),
			currentUser.ID,
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return([]models.Group{
			{
				ID:        1,
				Name:      "Test Group",
				CreatedBy: currentUser.ID,
			},
		}, nil).Times(1)
		groups, err := gs.GetAll(tx, *currentUser, queryParams)
		require.NoError(t, err)
		assert.NotNil(t, groups)
	})
}

func TestGetGroup(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	t.Run("Success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		gr.EXPECT().Get(gomock.Any(), int64(1)).Return(&models.Group{
			ID:        1,
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		}, nil).Times(1)
		group, err := gs.Get(tx, *currentUser, 1)
		require.NoError(t, err)
		assert.Equal(t, "Test Group", group.Name)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		group, err := gs.Get(tx, *currentUser, 1)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, group)
	})
}

func TestAddUsersToGroup(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	t.Run("Success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		gr.EXPECT().Get(gomock.Any(), int64(1)).Return(&models.Group{
			ID:   int64(1),
			Name: "Group name",
		}, nil).Times(1)
		user := &schemas.User{ID: int64(2)}
		ur.EXPECT().Get(gomock.Any(), user.ID).Return(&models.User{ID: user.ID}, nil).Times(1)
		gr.EXPECT().UserBelongsTo(gomock.Any(), int64(1), user.ID).Return(false, nil).Times(1)
		gr.EXPECT().AddUser(gomock.Any(), int64(1), user.ID).Return(nil).Times(1)
		err := gs.AddUsers(tx, *currentUser, int64(1), []int64{user.ID})
		require.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		groupID := int64(1) // Assuming the group ID is 1 for the test
		user := &schemas.User{ID: 3}
		err := gs.AddUsers(tx, *currentUser, groupID, []int64{user.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetGroupUsers(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	t.Run("Success", func(t *testing.T) {
		currentUser := &schemas.User{ID: int64(1), Role: types.UserRoleAdmin}
		groupID := int64(1)
		gr.EXPECT().Get(gomock.Any(), groupID).Return(&models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		}, nil).Times(1)
		user := &schemas.User{ID: int64(2)}
		gr.EXPECT().GetUsers(gomock.Any(), groupID).Return([]models.User{{ID: user.ID}}, nil).Times(1)
		users, err := gs.GetUsers(tx, *currentUser, groupID)
		require.NoError(t, err)
		assert.NotEmpty(t, users)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		groupID := int64(1)
		user := &schemas.User{ID: 3}
		err := gs.AddUsers(tx, *currentUser, groupID, []int64{user.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		users, err := gs.GetUsers(tx, *currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, users)
	})
}
func TestEditGroup(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	newGroupName := "Updated Group Name"
	t.Run("Success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		group := &models.Group{
			ID:        int64(1),
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		}
		gr.EXPECT().Get(gomock.Any(), group.ID).Return(group, nil).Times(1)
		editInfo := &schemas.EditGroup{
			Name: &newGroupName,
		}
		gr.EXPECT().Edit(gomock.Any(), group.ID, gomock.Any()).Return(&models.Group{
			ID:   group.ID,
			Name: newGroupName,
		}, nil).Times(1)
		updatedGroup, err := gs.Edit(tx, *currentUser, group.ID, editInfo)
		require.NoError(t, err)
		assert.Equal(t, newGroupName, updatedGroup.Name)
	})

	t.Run("Group not found", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		group := &models.Group{
			ID: int64(9999),
		}
		gr.EXPECT().Get(gomock.Any(), group.ID).Return(nil, errors.ErrGroupNotFound).Times(1)
		editInfo := &schemas.EditGroup{
			Name: &newGroupName,
		}
		updatedGroup, err := gs.Edit(tx, *currentUser, group.ID, editInfo)
		require.ErrorIs(t, err, errors.ErrGroupNotFound)
		assert.Nil(t, updatedGroup)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := &schemas.User{ID: 3, Role: types.UserRoleTeacher}
		group := &models.Group{
			ID:        int64(1),
			Name:      "Test Group",
			CreatedBy: 1, // Assuming the admin user ID is 1
		}
		gr.EXPECT().Get(gomock.Any(), group.ID).Return(group, nil).Times(1)
		editInfo := &schemas.EditGroup{
			Name: &newGroupName,
		}
		updatedGroup, err := gs.Edit(tx, *currentUser, group.ID, editInfo)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, updatedGroup)
	})

	t.Run("Validation error", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		group := &models.Group{
			ID: int64(1),
		}
		editInfo := &schemas.EditGroup{
			Name: nil, // Invalid as name is required
		}
		updatedGroup, err := gs.Edit(tx, *currentUser, group.ID, editInfo)
		require.Error(t, err)
		assert.Nil(t, updatedGroup)
	})
}
func TestDeleteUsersFromGroup(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	t.Run("Success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		groupID := int64(1)
		gr.EXPECT().Get(gomock.Any(), groupID).Return(&models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		}, nil).Times(1)
		user := &schemas.User{ID: 2}
		gr.EXPECT().UserBelongsTo(gomock.Any(), groupID, user.ID).Return(true, nil).Times(1)
		gr.EXPECT().DeleteUser(gomock.Any(), groupID, user.ID).Return(nil).Times(1)
		ur.EXPECT().Get(gomock.Any(), user.ID).Return(&models.User{ID: user.ID}, nil).Times(1)
		err := gs.DeleteUsers(tx, *currentUser, groupID, []int64{user.ID})
		require.NoError(t, err)
	})

	t.Run("Not authorized student", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		groupID := int64(1) // Assuming the group ID is 1 for the test

		user := &schemas.User{ID: 3}
		err := gs.DeleteUsers(tx, *currentUser, groupID, []int64{user.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		currentUser := &schemas.User{ID: 3, Role: types.UserRoleTeacher}
		group := &models.Group{
			ID:        int64(1),
			Name:      "Test Group",
			CreatedBy: 1, // Assuming the admin user ID is 1
		}
		gr.EXPECT().Get(gomock.Any(), group.ID).Return(group, nil).Times(1)
		user := &schemas.User{ID: 2}
		err := gs.DeleteUsers(tx, *currentUser, group.ID, []int64{user.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("User not found", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		groupID := int64(1)
		gr.EXPECT().Get(gomock.Any(), groupID).Return(&models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		}, nil).Times(1)
		ur.EXPECT().Get(gomock.Any(), int64(9999)).Return(nil, errors.ErrUserNotFound).Times(1)
		err := gs.DeleteUsers(tx, *currentUser, groupID, []int64{9999})
		require.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("Group not found", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		gr.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.ErrNotFound).Times(1)
		err := gs.DeleteUsers(tx, *currentUser, int64(9999), []int64{currentUser.ID})
		require.ErrorIs(t, err, errors.ErrNotFound)
	})
}
func TestGetGroupTasks(t *testing.T) {
	tx := database.NewDB(&gorm.DB{})
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	t.Run("Success for Admin", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		groupID := int64(1)
		gr.EXPECT().Get(gomock.Any(), groupID).Return(&models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		}, nil).Times(1)
		gr.EXPECT().GetTasks(gomock.Any(), groupID).Return([]models.Task{}, nil).Times(1)
		tasks, err := gs.GetTasks(tx, *currentUser, groupID)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
	})

	t.Run("Success for Teacher", func(t *testing.T) {
		currentUser := &schemas.User{ID: 3, Role: types.UserRoleTeacher}
		groupID := int64(1)
		gr.EXPECT().Get(gomock.Any(), groupID).Return(&models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		}, nil).Times(1)
		gr.EXPECT().GetTasks(gomock.Any(), groupID).Return([]models.Task{}, nil).Times(1)
		tasks, err := gs.GetTasks(tx, *currentUser, groupID)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
	})

	t.Run("Not authorized for Teacher", func(t *testing.T) {
		currentUser := &schemas.User{ID: 3, Role: types.UserRoleTeacher}
		group := &models.Group{
			ID:        int64(1),
			Name:      "Test Group",
			CreatedBy: 1, // Assuming the admin user ID is 1
		}
		gr.EXPECT().Get(gomock.Any(), group.ID).Return(group, nil).Times(1)
		tasks, err := gs.GetTasks(tx, *currentUser, group.ID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, tasks)
	})

	t.Run("Success for Student", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		adminUser := &schemas.User{ID: 1}
		groupID := int64(1)
		gr.EXPECT().Get(gomock.Any(), groupID).Return(&models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedBy: adminUser.ID,
		}, nil).Times(1)
		gr.EXPECT().UserBelongsTo(gomock.Any(), groupID, currentUser.ID).Return(true, nil).Times(1)
		gr.EXPECT().GetTasks(gomock.Any(), groupID).Return([]models.Task{}, nil).Times(1)
		tasks, err := gs.GetTasks(tx, *currentUser, groupID)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
	})

	t.Run("Not authorized for Student", func(t *testing.T) {
		currentUser := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		groupID := int64(1) // Assuming the group ID is 1 for the test
		gr.EXPECT().Get(gomock.Any(), groupID).Return(&models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedBy: 1, // Assuming the admin user ID is 1
		}, nil).Times(1)
		gr.EXPECT().UserBelongsTo(gomock.Any(), groupID, currentUser.ID).Return(false, nil).Times(1)
		tasks, err := gs.GetTasks(tx, *currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, tasks)
	})

	t.Run("Group not found", func(t *testing.T) {
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		groupID := int64(9999)
		gr.EXPECT().Get(gomock.Any(), groupID).Return(nil, errors.ErrGroupNotFound).Times(1)
		tasks, err := gs.GetTasks(tx, *currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrGroupNotFound)
		assert.Nil(t, tasks)
	})
}
