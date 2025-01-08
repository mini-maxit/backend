package service

import (
	"fmt"
	"testing"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type taskServiceTest struct {
	tx          *gorm.DB
	config      *config.Config
	ur          repository.UserRepository
	tr          repository.TaskRepository
	taskService TaskService
	counter     int64
}

func newTaskServiceTest() *taskServiceTest {
	tx := &gorm.DB{}
	config := testutils.NewTestConfig()
	ur := testutils.NewMockUserRepository()
	tr := testutils.NewMockTaskRepository()
	ts := NewTaskService(config, tr)

	return &taskServiceTest{
		tx:          tx,
		config:      config,
		ur:          ur,
		tr:          tr,
		taskService: ts,
	}
}

func (tst *taskServiceTest) createUser(t *testing.T) int64 {
	tst.counter++
	userId, err := tst.ur.CreateUser(tst.tx, &models.User{
		Name:         fmt.Sprintf("Test User %d", tst.counter),
		Surname:      fmt.Sprintf("Test Surname %d", tst.counter),
		Email:        fmt.Sprintf("email%d@email.com", tst.counter),
		Username:     fmt.Sprintf("testuser%d", tst.counter),
		PasswordHash: "password",
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return userId
}

func TestCreateTask(t *testing.T) {
	tst := newTaskServiceTest()

	t.Run("Success", func(t *testing.T) {
		userId := tst.createUser(t)
		taskId, err := tst.taskService.Create(tst.tx, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: userId,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
	})

	// We want to have clear task repository for this state, and this is the quickest way
	tst = newTaskServiceTest()
	t.Run("Non unique title", func(t *testing.T) {
		userId := tst.createUser(t)
		taskId, err := tst.taskService.Create(tst.tx, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: userId,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), taskId)
		taskId, err = tst.taskService.Create(tst.tx, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: userId,
		})
		assert.ErrorIs(t, err, ErrTaskExists)
		assert.Equal(t, int64(0), taskId)
	})
}

func TestGetTaskByTitle(t *testing.T) {
	tst := newTaskServiceTest()
	t.Run("Success", func(t *testing.T) {
		userId := tst.createUser(t)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: userId,
		}
		taskId, err := tst.taskService.Create(tst.tx, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		taskResp, err := tst.taskService.GetTaskByTitle(tst.tx, task.Title)
		assert.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		task, err := tst.taskService.GetTaskByTitle(tst.tx, "Nonexistent Task")
		assert.ErrorIs(t, err, ErrTaskNotFound)
		assert.Nil(t, task)
	})
}

func TestGetAllTasks(t *testing.T) {
	tst := newTaskServiceTest()
	queryParams := map[string]string{"limit": "10", "offset": "0", "sort": "id:asc"}
	t.Run("No tasks", func(t *testing.T) {
		tasks, err := tst.taskService.GetAll(tst.tx, queryParams)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		userId := tst.createUser(t)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: userId,
		}
		taskId, err := tst.taskService.Create(tst.tx, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		tasks, err := tst.taskService.GetAll(tst.tx, queryParams)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})
}

func TestGetTask(t *testing.T) {
	tst := newTaskServiceTest()

	t.Run("Success", func(t *testing.T) {
		userId := tst.createUser(t)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: userId,
		}
		taskId, err := tst.taskService.Create(tst.tx, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		taskResp, err := tst.taskService.GetTask(tst.tx, taskId)
		assert.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})
}

func TestUpdateTask(t *testing.T) {
	tst := newTaskServiceTest()
	t.Run("Success", func(t *testing.T) {
		userId := tst.createUser(t)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: userId,
		}
		taskId, err := tst.taskService.Create(tst.tx, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		updatedTask := schemas.UpdateTask{
			Title: "Updated Task",
		}
		err = tst.taskService.UpdateTask(tst.tx, taskId, updatedTask)
		assert.NoError(t, err)
		taskResp, err := tst.taskService.GetTask(tst.tx, taskId)
		assert.NoError(t, err)
		assert.Equal(t, updatedTask.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})
	t.Run("Nonexistent task", func(t *testing.T) {
		updatedTask := schemas.UpdateTask{
			Title: "Updated Task",
		}
		err := tst.taskService.UpdateTask(tst.tx, 0, updatedTask)
		assert.ErrorIs(t, err, ErrTaskNotFound)
	})
}
