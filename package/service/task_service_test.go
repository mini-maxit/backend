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

type taskServiceTest struct {
	tx          *gorm.DB
	config      *config.Config
	ur          repository.UserRepository
	tr          repository.TaskRepository
	sr          repository.SubmissionRepository
	taskService TaskService
	savePoint   string
}

func newTaskServiceTest(t *testing.T) *taskServiceTest {
	tx := testutils.NewTestTx(t)
	config := testutils.NewTestConfig()
	ur, err := repository.NewUserRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	tr, err := repository.NewTaskRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	sr, err := repository.NewSubmissionRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	ts := NewTaskService(config, tr, sr)
	savePoint := "savepoint"
	tx.SavePoint(savePoint)
	return &taskServiceTest{
		tx:          tx,
		config:      config,
		ur:          ur,
		tr:          tr,
		sr:          sr,
		taskService: ts,
		savePoint:   savePoint,
	}
}

func (tst *taskServiceTest) createUser(t *testing.T) int64 {
	userId, err := tst.ur.CreateUser(tst.tx, &models.User{
		Name:         "Test User",
		Surname:      "Test Surname",
		Email:        "email@email.com",
		Username:     "testuser",
		PasswordHash: "password",
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return userId
}

func (tst *taskServiceTest) rollbackToSavePoint() {
	tst.tx.RollbackTo(tst.savePoint)
}

func TestCreateTask(t *testing.T) {
	tst := newTaskServiceTest(t)

	t.Run("Success", func(t *testing.T) {
		userId := tst.createUser(t)
		taskId, err := tst.taskService.Create(tst.tx, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: userId,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		tst.rollbackToSavePoint()
	})

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
		tst.rollbackToSavePoint()
	})
	tst.tx.Rollback()
}

func TestGetTaskByTitle(t *testing.T) {
	tst := newTaskServiceTest(t)
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
		tst.rollbackToSavePoint()
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		task, err := tst.taskService.GetTaskByTitle(tst.tx, "Nonexistent Task")
		assert.ErrorIs(t, err, ErrTaskNotFound)
		assert.Nil(t, task)
		tst.rollbackToSavePoint()
	})
	tst.tx.Rollback()
}

func TestGetAllTasks(t *testing.T) {
	tst := newTaskServiceTest(t)
	queryParams := map[string]string{"limit": "10", "offset": "0", "sort": "id:asc"}
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
		tst.rollbackToSavePoint()
	})

	t.Run("No tasks", func(t *testing.T) {
		tasks, err := tst.taskService.GetAll(tst.tx, queryParams)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
		tst.rollbackToSavePoint()
	})

	tst.tx.Rollback()
}

func TestGetTask(t *testing.T) {
	tst := newTaskServiceTest(t)

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
		tst.rollbackToSavePoint()
	})
	tst.tx.Rollback()
}

func TestUpdateTask(t *testing.T) {
	tst := newTaskServiceTest(t)

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
		tst.rollbackToSavePoint()
	})
	t.Run("Nonexistent task", func(t *testing.T) {
		updatedTask := schemas.UpdateTask{
			Title: "Updated Task",
		}
		err := tst.taskService.UpdateTask(tst.tx, 0, updatedTask)
		assert.ErrorIs(t, err, ErrTaskNotFound)
		tst.rollbackToSavePoint()
	})
	tst.tx.Rollback()
}
