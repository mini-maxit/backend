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

type taskServiceTest struct {
	tx          *gorm.DB
	ur          repository.UserRepository
	tr          repository.TaskRepository
	gr          repository.GroupRepository
	taskService TaskService
	counter     int64
}

func newTaskServiceTest() *taskServiceTest {
	tx := &gorm.DB{}
	config := testutils.NewTestConfig()
	ur := testutils.NewMockUserRepository()
	gr := testutils.NewMockGroupRepository(ur)
	tr := testutils.NewMockTaskRepository(gr)
	ts := NewTaskService(config.FileStorageUrl, tr, nil, ur, gr)

	return &taskServiceTest{
		tx:          tx,
		ur:          ur,
		tr:          tr,
		gr:          gr,
		taskService: ts,
	}
}

func (tst *taskServiceTest) createUser(t *testing.T, role types.UserRole) schemas.User {
	tst.counter++
	userId, err := tst.ur.CreateUser(tst.tx, &models.User{
		Name:         fmt.Sprintf("Test User %d", tst.counter),
		Surname:      fmt.Sprintf("Test Surname %d", tst.counter),
		Email:        fmt.Sprintf("email%d@email.com", tst.counter),
		Username:     fmt.Sprintf("testuser%d", tst.counter),
		Role:         role,
		PasswordHash: "password",
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	user_model, err := tst.ur.GetUser(tst.tx, userId)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	user := schemas.User{
		Id:   user_model.Id,
		Role: user_model.Role,
	}
	return user
}

func TestCreateTask(t *testing.T) {
	tst := newTaskServiceTest()

	t.Run("Success", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		taskId, err := tst.taskService.Create(tst.tx, current_user, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
	})

	// We want to have clear task repository for this state, and this is the quickest way
	tst = newTaskServiceTest()
	t.Run("Non unique title", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		taskId, err := tst.taskService.Create(tst.tx, current_user, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: current_user.Id,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), taskId)
		taskId, err = tst.taskService.Create(tst.tx, current_user, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: current_user.Id,
		})
		assert.ErrorIs(t, err, errors.ErrTaskExists)
		assert.Equal(t, int64(0), taskId)
	})

	t.Run("Not authorized", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleStudent)
		taskId, err := tst.taskService.Create(tst.tx, current_user, &schemas.Task{
			Title:     "Test Student Task",
			CreatedBy: current_user.Id,
		})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Equal(t, int64(0), taskId)
	})
}

func TestGetTaskByTitle(t *testing.T) {
	tst := newTaskServiceTest()
	t.Run("Success", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: current_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, current_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		taskResp, err := tst.taskService.GetTaskByTitle(tst.tx, task.Title)
		assert.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		task, err := tst.taskService.GetTaskByTitle(tst.tx, "Nonexistent Task")
		assert.ErrorIs(t, err, errors.ErrTaskNotFound)
		assert.Nil(t, task)
	})
}

func TestGetAllTasks(t *testing.T) {
	tst := newTaskServiceTest()
	queryParams := map[string]interface{}{"limit": uint64(10), "offset": uint64(0), "sort": "id:asc"}
	t.Run("No tasks", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		tasks, err := tst.taskService.GetAll(tst.tx, current_user, queryParams)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: current_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, current_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		tasks, err := tst.taskService.GetAll(tst.tx, current_user, queryParams)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})

	t.Run("Not authorized", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleStudent)
		tasks, err := tst.taskService.GetAll(tst.tx, current_user, queryParams)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})
}

func TestGetTask(t *testing.T) {
	tst := newTaskServiceTest()
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: admin_user.Id,
	}

	taskId, err := tst.taskService.Create(tst.tx, admin_user, task)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Run("Success", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		taskResp, err := tst.taskService.GetTask(tst.tx, current_user, taskId)
		assert.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Sucess with assigned task to user", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		err = tst.taskService.AssignTaskToUsers(tst.tx, admin_user, taskId, []int64{student_user.Id})
		assert.NoError(t, err)
		taskResp, err := tst.taskService.GetTask(tst.tx, student_user, taskId)
		fmt.Printf("Error: %v\n", err)
		assert.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Sucess with assigned task to group", func(t *testing.T) {
		group_model := &models.Group{
			Name: "Test Group",
		}
		groupId, err := tst.gr.CreateGroup(tst.tx, group_model)
		assert.NoError(t, err)
		err = tst.taskService.AssignTaskToGroups(tst.tx, admin_user, taskId, []int64{groupId})
		assert.NoError(t, err)
		student_user := tst.createUser(t, types.UserRoleStudent)
		err = tst.gr.AddUserToGroup(tst.tx, groupId, student_user.Id)
		assert.NoError(t, err)
		taskResp, err := tst.taskService.GetTask(tst.tx, student_user, taskId)
		assert.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Success with created task", func(t *testing.T) {
		teacher_user := tst.createUser(t, types.UserRoleTeacher)
		task := &schemas.Task{
			Title:     "Test Task 2",
			CreatedBy: teacher_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, teacher_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		taskResp, err := tst.taskService.GetTask(tst.tx, teacher_user, taskId)
		assert.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Not authorized", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		taskResp, err := tst.taskService.GetTask(tst.tx, student_user, taskId)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskResp)
	})
}

func TestAssignTaskToUsers(t *testing.T) {
	tst := newTaskServiceTest()
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: admin_user.Id,
	}

	taskId, err := tst.taskService.Create(tst.tx, admin_user, task)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Run("Success", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		err := tst.taskService.AssignTaskToUsers(tst.tx, admin_user, taskId, []int64{student_user.Id})
		assert.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		err := tst.taskService.AssignTaskToUsers(tst.tx, admin_user, 0, []int64{student_user.Id})
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		err := tst.taskService.AssignTaskToUsers(tst.tx, student_user, taskId, []int64{student_user.Id})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestAssignTaskToGroups(t *testing.T) {
	tst := newTaskServiceTest()
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: admin_user.Id,
	}

	taskId, err := tst.taskService.Create(tst.tx, admin_user, task)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Run("Success", func(t *testing.T) {
		group_model := &models.Group{
			Name: "Test Group",
		}
		groupId, err := tst.gr.CreateGroup(tst.tx, group_model)
		assert.NoError(t, err)
		err = tst.taskService.AssignTaskToGroups(tst.tx, admin_user, taskId, []int64{groupId})
		assert.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		group_model := &models.Group{
			Name: "Test Group",
		}
		groupId, err := tst.gr.CreateGroup(tst.tx, group_model)
		assert.NoError(t, err)
		err = tst.taskService.AssignTaskToGroups(tst.tx, admin_user, 0, []int64{groupId})
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		group_model := &models.Group{
			Name: "Test Group",
		}
		groupId, err := tst.gr.CreateGroup(tst.tx, group_model)
		assert.NoError(t, err)
		student_user := tst.createUser(t, types.UserRoleStudent)
		err = tst.gr.AddUserToGroup(tst.tx, groupId, student_user.Id)
		assert.NoError(t, err)
		err = tst.taskService.AssignTaskToGroups(tst.tx, student_user, taskId, []int64{groupId})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetAllAssignedTasks(t *testing.T) {
	tst := newTaskServiceTest()
	query_params := map[string]interface{}{"limit": uint64(10), "offset": uint64(0), "sort": "id:asc"}
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: admin_user.Id,
	}

	taskId, err := tst.taskService.Create(tst.tx, admin_user, task)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Run("No tasks", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		tasks, err := tst.taskService.GetAllAssignedTasks(tst.tx, student_user, query_params)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		err := tst.taskService.AssignTaskToUsers(tst.tx, admin_user, taskId, []int64{student_user.Id})
		assert.NoError(t, err)
		group := &models.Group{
			Name: "Test Group",
		}
		groupId, err := tst.gr.CreateGroup(tst.tx, group)
		assert.NoError(t, err)
		err = tst.gr.AddUserToGroup(tst.tx, groupId, student_user.Id)
		assert.NoError(t, err)
		err = tst.taskService.AssignTaskToGroups(tst.tx, admin_user, taskId, []int64{groupId})
		assert.NoError(t, err)
		tasks, err := tst.taskService.GetAllAssignedTasks(tst.tx, student_user, query_params)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, 2, len(tasks))
	})
}

func TestDeleteTask(t *testing.T) {
	tst := newTaskServiceTest()
	t.Run("Success", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: current_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, current_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		err = tst.taskService.DeleteTask(tst.tx, current_user, taskId)
		assert.NoError(t, err)
		_, err = tst.taskService.GetTask(tst.tx, current_user, taskId)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		err := tst.taskService.DeleteTask(tst.tx, current_user, 0)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleStudent)
		admin_user := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: current_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, admin_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		err = tst.taskService.DeleteTask(tst.tx, current_user, taskId)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}
func TestUpdateTask(t *testing.T) {
	tst := newTaskServiceTest()
	admin_user := tst.createUser(t, types.UserRoleAdmin)

	t.Run("Success", func(t *testing.T) {
		current_user := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: current_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, current_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		newTitle := "Updated Task"
		updatedTask := &schemas.EditTask{Title: &newTitle}
		err = tst.taskService.EditTask(tst.tx, admin_user, taskId, updatedTask)
		assert.NoError(t, err)
		taskResp, err := tst.taskService.GetTask(tst.tx, current_user, taskId)
		assert.NoError(t, err)
		assert.Equal(t, *updatedTask.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})
	t.Run("Nonexistent task", func(t *testing.T) {
		newTitle := "Updated Task"
		updatedTask := &schemas.EditTask{Title: &newTitle}
		err := tst.taskService.EditTask(tst.tx, admin_user, 0, updatedTask)
		assert.ErrorIs(t, err, errors.ErrTaskNotFound)
	})
}
