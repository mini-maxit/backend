package service

import (
	"fmt"
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type groupServiceTest struct {
	tx           *gorm.DB
	ur           repository.UserRepository
	gr           repository.GroupRepository
	groupService GroupService
	counter      int64
}

func newGroupServiceTest() *groupServiceTest {
	tx := &gorm.DB{}
	ur := testutils.NewMockUserRepository()
	gr := testutils.NewMockGroupRepository(ur)
	gs := NewGroupService(gr, ur, NewUserService(ur))

	return &groupServiceTest{
		tx:           tx,
		ur:           ur,
		gr:           gr,
		groupService: gs,
	}
}

func (gst *groupServiceTest) createUser(t *testing.T, role types.UserRole) schemas.User {
	gst.counter++
	userId, err := gst.ur.CreateUser(gst.tx, &models.User{
		Name:         fmt.Sprintf("Test User %d", gst.counter),
		Surname:      fmt.Sprintf("Test Surname %d", gst.counter),
		Email:        fmt.Sprintf("email%d@email.com", gst.counter),
		Username:     fmt.Sprintf("testuser%d", gst.counter),
		Role:         role,
		PasswordHash: "password",
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	user_model, err := gst.ur.GetUser(gst.tx, userId)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	user := schemas.User{
		Id:   user_model.Id,
		Role: user_model.Role,
	}
	return user
}

func TestCreateGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleAdmin)
		groupId, err := gst.groupService.CreateGroup(gst.tx, current_user, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
	})

	t.Run("Not authorized", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleStudent)
		groupId, err := gst.groupService.CreateGroup(gst.tx, current_user, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Equal(t, int64(0), groupId)
	})
}

func TestDeleteGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleAdmin)
		groupId, err := gst.groupService.CreateGroup(gst.tx, current_user, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		err = gst.groupService.DeleteGroup(gst.tx, current_user, groupId)
		assert.NoError(t, err)
	})

	t.Run("Not authorized student", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleStudent)
		groupId, err := gst.groupService.CreateGroup(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		err = gst.groupService.DeleteGroup(gst.tx, current_user, groupId)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleTeacher)
		groupId, err := gst.groupService.CreateGroup(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: gst.createUser(t, types.UserRoleStudent).Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		err = gst.groupService.DeleteGroup(gst.tx, current_user, groupId)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetAllGroup(t *testing.T) {
	gst := newGroupServiceTest()

	queryParams := map[string]interface{}{"limit": uint64(10), "offset": uint64(0), "sort": "id:asc"}
	t.Run("No groups", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleAdmin)
		groups, err := gst.groupService.GetAllGroup(gst.tx, current_user, queryParams)
		assert.NoError(t, err)
		assert.Empty(t, groups)
	})

	t.Run("Success", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleAdmin)
		groupId, err := gst.groupService.CreateGroup(gst.tx, current_user, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		groups, err := gst.groupService.GetAllGroup(gst.tx, current_user, queryParams)
		assert.NoError(t, err)
		assert.NotEmpty(t, groups)
	})

	t.Run("Not authorized", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleStudent)
		groups, err := gst.groupService.GetAllGroup(gst.tx, current_user, queryParams)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, groups)
	})
}

func TestGetGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleAdmin)
		groupId, err := gst.groupService.CreateGroup(gst.tx, current_user, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		group, err := gst.groupService.GetGroup(gst.tx, current_user, groupId)
		assert.NoError(t, err)
		assert.Equal(t, "Test Group", group.Name)
	})

	t.Run("Not authorized", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleStudent)
		groupId, err := gst.groupService.CreateGroup(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		group, err := gst.groupService.GetGroup(gst.tx, current_user, groupId)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, group)
	})
}

func TestAddUsersToGroup(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleAdmin)
		groupId, err := gst.groupService.CreateGroup(gst.tx, current_user, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsersToGroup(gst.tx, current_user, groupId, []int64{user.Id})
		assert.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleStudent)
		groupId, err := gst.groupService.CreateGroup(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsersToGroup(gst.tx, current_user, groupId, []int64{user.Id})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetGroupUsers(t *testing.T) {
	gst := newGroupServiceTest()

	t.Run("Success", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleAdmin)
		groupId, err := gst.groupService.CreateGroup(gst.tx, current_user, &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsersToGroup(gst.tx, current_user, groupId, []int64{user.Id})
		assert.NoError(t, err)
		users, err := gst.groupService.GetGroupUsers(gst.tx, current_user, groupId)
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
	})

	t.Run("Not authorized", func(t *testing.T) {
		current_user := gst.createUser(t, types.UserRoleStudent)
		groupId, err := gst.groupService.CreateGroup(gst.tx, gst.createUser(t, types.UserRoleAdmin), &schemas.Group{
			Name:      "Test Group",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, groupId)
		user := gst.createUser(t, types.UserRoleStudent)
		err = gst.groupService.AddUsersToGroup(gst.tx, gst.createUser(t, types.UserRoleAdmin), groupId, []int64{user.Id})
		assert.NoError(t, err)
		users, err := gst.groupService.GetGroupUsers(gst.tx, current_user, groupId)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Nil(t, users)
	})
}
