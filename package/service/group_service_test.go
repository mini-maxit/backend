package service_test

import (
	"fmt"
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type groupServiceTest struct {
	tx           *gorm.DB
	ur           repository.UserRepository
	gr           repository.GroupRepository
	groupService service.GroupService
	counter      int64
}

func newGroupServiceTest() *groupServiceTest {
	tx := &gorm.DB{}
	ur := testutils.NewMockUserRepository()
	gr := testutils.NewMockGroupRepository(ur)
	gs := service.NewGroupService(gr, ur, service.NewUserService(ur))

	return &groupServiceTest{
		tx:           tx,
		ur:           ur,
		gr:           gr,
		groupService: gs,
	}
}

func (gst *groupServiceTest) createUser(t *testing.T, role types.UserRole) schemas.User {
	gst.counter++
	userID, err := gst.ur.Create(gst.tx, &models.User{
		Name:         fmt.Sprintf("Test User %d", gst.counter),
		Surname:      fmt.Sprintf("Test Surname %d", gst.counter),
		Email:        fmt.Sprintf("email%d@email.com", gst.counter),
		Username:     fmt.Sprintf("testuser%d", gst.counter),
		Role:         role,
		PasswordHash: "password",
	})
	require.NoError(t, err)

	userModel, err := gst.ur.Get(gst.tx, userID)
	require.NoError(t, err)

	user := schemas.User{
		ID:   userModel.ID,
		Role: userModel.Role,
	}
	return user
}

func TestCreateGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Equal(t, int64(0), groupID)
	})
}

func TestDeleteGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		err = gst.groupService.Delete(gst.tx, currentUser, groupID)
		require.NoError(t, err)
	})

	t.Run("Not authorized student", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		err = gst.groupService.Delete(gst.tx, currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleTeacher)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: gst.createUser(t, types.UserRoleStudent).ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		err = gst.groupService.Delete(gst.tx, currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetAllGroup(t *testing.T) {
	gst := newGroupServiceTest()

	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}
	t.Run("No groups", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groups, err := gst.groupService.GetAll(gst.tx, currentUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, groups)
	})

	t.Run("Success", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		groups, err := gst.groupService.GetAll(gst.tx, currentUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, groups)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groups, err := gst.groupService.GetAll(gst.tx, currentUser, queryParams)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, groups)
	})

	t.Run("Success for teacher", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleTeacher)
		groups, err := gst.groupService.GetAll(gst.tx, currentUser, queryParams)
		require.NoError(t, err)
		assert.NotNil(t, groups)
	})
}

func TestGetGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		group, err := gst.groupService.Get(gst.tx, currentUser, groupID)
		require.NoError(t, err)
		assert.Equal(t, "Test Group", group.Name)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		group, err := gst.groupService.Get(gst.tx, currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, group)
	})
}

func TestAddUsersToGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsers(gst.tx, currentUser, groupID, []int64{user.ID})
		require.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsers(gst.tx, currentUser, groupID, []int64{user.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetGroupUsers(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsers(gst.tx, currentUser, groupID, []int64{user.ID})
		require.NoError(t, err)
		users, err := gst.groupService.GetUsers(gst.tx, currentUser, groupID)
		require.NoError(t, err)
		assert.NotEmpty(t, users)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)
		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsers(gst.tx, gst.createUser(t, types.UserRoleAdmin), groupID, []int64{user.ID})
		require.NoError(t, err)
		users, err := gst.groupService.GetUsers(gst.tx, currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, users)
	})
}
func TestEditGroup(t *testing.T) {
	gst := newGroupServiceTest()
	newGroupName := "Updated Group Name"
	t.Run("Success", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		editInfo := &schemas.EditGroup{
			Name: &newGroupName,
		}
		updatedGroup, err := gst.groupService.Edit(gst.tx, currentUser, groupID, editInfo)
		require.NoError(t, err)
		assert.Equal(t, newGroupName, updatedGroup.Name)
	})

	t.Run("Group not found", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		editInfo := &schemas.EditGroup{
			Name: &newGroupName,
		}
		updatedGroup, err := gst.groupService.Edit(gst.tx, currentUser, 9999, editInfo)
		require.ErrorIs(t, err, errors.ErrGroupNotFound)
		assert.Nil(t, updatedGroup)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleTeacher)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: gst.createUser(t, types.UserRoleAdmin).ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		editInfo := &schemas.EditGroup{
			Name: &newGroupName,
		}
		updatedGroup, err := gst.groupService.Edit(gst.tx, currentUser, groupID, editInfo)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, updatedGroup)
	})

	t.Run("Validation error", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		editInfo := &schemas.EditGroup{
			Name: nil, // Invalid as name is required
		}
		updatedGroup, err := gst.groupService.Edit(gst.tx, currentUser, groupID, editInfo)
		require.Error(t, err)
		assert.Nil(t, updatedGroup)
	})
}
func TestDeleteUsersFromGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsers(gst.tx, currentUser, groupID, []int64{user.ID})
		require.NoError(t, err)

		err = gst.groupService.DeleteUsers(gst.tx, currentUser, groupID, []int64{user.ID})
		require.NoError(t, err)
	})

	t.Run("Not authorized student", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsers(gst.tx, gst.createUser(t, types.UserRoleAdmin), groupID, []int64{user.ID})
		require.NoError(t, err)

		err = gst.groupService.DeleteUsers(gst.tx, currentUser, groupID, []int64{user.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleTeacher)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: gst.createUser(t, types.UserRoleAdmin).ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsers(gst.tx, gst.createUser(t, types.UserRoleAdmin), groupID, []int64{user.ID})
		require.NoError(t, err)

		err = gst.groupService.DeleteUsers(gst.tx, currentUser, groupID, []int64{user.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("User not found", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		err = gst.groupService.DeleteUsers(gst.tx, currentUser, groupID, []int64{9999})
		require.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("Group not found", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		err := gst.groupService.DeleteUsers(gst.tx, currentUser, 9999, []int64{currentUser.ID})
		require.ErrorIs(t, err, errors.ErrNotFound)
	})
}
func TestGetGroupTasks(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success for Admin", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		tasks, err := gst.groupService.GetTasks(gst.tx, currentUser, groupID)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
	})

	t.Run("Success for Teacher", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleTeacher)
		groupID, err := gst.groupService.Create(gst.tx, currentUser, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		tasks, err := gst.groupService.GetTasks(gst.tx, currentUser, groupID)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
	})

	t.Run("Not authorized for Teacher", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleTeacher)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: gst.createUser(t, types.UserRoleAdmin).ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		tasks, err := gst.groupService.GetTasks(gst.tx, currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, tasks)
	})

	t.Run("Success for Student", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: gst.createUser(t, types.UserRoleAdmin).ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		err = gst.groupService.AddUsers(gst.tx, gst.createUser(t, types.UserRoleAdmin), groupID, []int64{currentUser.ID})
		require.NoError(t, err)

		tasks, err := gst.groupService.GetTasks(gst.tx, currentUser, groupID)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
	})

	t.Run("Not authorized for Student", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleStudent)
		groupID, err := gst.groupService.Create(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: gst.createUser(t, types.UserRoleAdmin).ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, groupID)

		tasks, err := gst.groupService.GetTasks(gst.tx, currentUser, groupID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, tasks)
	})

	t.Run("Group not found", func(t *testing.T) {
		currentUser := gst.createUser(t, types.UserRoleAdmin)
		tasks, err := gst.groupService.GetTasks(gst.tx, currentUser, 9999)
		require.Error(t, err)
		assert.Nil(t, tasks)
	})
}
