package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type contestServiceTest struct {
	tx             *gorm.DB
	ur             repository.UserRepository
	tr             repository.TaskRepository
	cr             repository.ContestRepository
	contestService ContestService
	counter        int64
}

func newContestServiceTest() *contestServiceTest {
	tx := &gorm.DB{}
	ur := testutils.NewMockUserRepository()
	gr := testutils.NewMockGroupRepository(ur)
	tr := testutils.NewMockTaskRepository(gr)
	cr := testutils.NewMockContestRepository(tr)
	cs := NewContestService(cr, tr, ur)

	return &contestServiceTest{
		tx:             tx,
		ur:             ur,
		tr:             tr,
		cr:             cr,
		contestService: cs,
	}
}

func (cst *contestServiceTest) createUser(t *testing.T, role types.UserRole) schemas.User {
	cst.counter++
	userId, err := cst.ur.CreateUser(cst.tx, &models.User{
		Name:         fmt.Sprintf("Test User %d", cst.counter),
		Surname:      fmt.Sprintf("Test Surname %d", cst.counter),
		Email:        fmt.Sprintf("email%d@email.com", cst.counter),
		Username:     fmt.Sprintf("testuser%d", cst.counter),
		Role:         role,
		PasswordHash: "password",
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	user_model, err := cst.ur.GetUser(cst.tx, userId)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	user := schemas.User{
		Id:   user_model.Id,
		Role: user_model.Role,
	}
	return user
}

func (cst *contestServiceTest) createTask(t *testing.T, currentUser schemas.User, title string) int64 {
	taskModel := &models.Task{
		Title:     title,
		CreatedBy: currentUser.Id,
	}
	taskId, err := cst.tr.Create(cst.tx, taskModel)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return taskId
}

func TestCreateContest(t *testing.T) {
	cst := newContestServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, contestId)
	})

	cst = newContestServiceTest()
	t.Run("Non unique title", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), contestId)
		
		contestId, err = cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description 2",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.ErrorIs(t, err, errors.ErrContestExists)
		assert.Equal(t, int64(0), contestId)
	})

	t.Run("Invalid time range", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(-24 * time.Hour) // End before start
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Invalid Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.ErrorIs(t, err, errors.ErrInvalidTimeRange)
		assert.Equal(t, int64(0), contestId)
	})

	t.Run("Not authorized - student", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleStudent)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Unauthorized Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Equal(t, int64(0), contestId)
	})

	t.Run("Success - teacher", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleTeacher)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Teacher Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), contestId)
	})
}

func TestEditContest(t *testing.T) {
	cst := newContestServiceTest()

	t.Run("Success - update title", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Original Contest",
			Description: "Original Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		newTitle := "Updated Contest"
		err = cst.contestService.EditContest(cst.tx, currentUser, contestId, &schemas.EditContest{
			Title: &newTitle,
		})
		assert.NoError(t, err)

		contest, err := cst.contestService.GetContest(cst.tx, currentUser, contestId)
		assert.NoError(t, err)
		assert.Equal(t, newTitle, contest.Title)
	})

	cst = newContestServiceTest()
	t.Run("Success - update description", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Original Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		newDescription := "Updated Description"
		err = cst.contestService.EditContest(cst.tx, currentUser, contestId, &schemas.EditContest{
			Description: &newDescription,
		})
		assert.NoError(t, err)

		contest, err := cst.contestService.GetContest(cst.tx, currentUser, contestId)
		assert.NoError(t, err)
		assert.Equal(t, newDescription, contest.Description)
	})

	t.Run("Contest not found", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		newTitle := "Updated Contest"
		err := cst.contestService.EditContest(cst.tx, currentUser, 999, &schemas.EditContest{
			Title: &newTitle,
		})
		assert.ErrorIs(t, err, errors.ErrContestNotFound)
	})

	t.Run("Not authorized - teacher editing another's contest", func(t *testing.T) {
		adminUser := cst.createUser(t, types.UserRoleAdmin)
		teacherUser := cst.createUser(t, types.UserRoleTeacher)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, adminUser, &schemas.CreateContest{
			Title:       "Admin Contest",
			Description: "Admin Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		newTitle := "Hacked Contest"
		err = cst.contestService.EditContest(cst.tx, teacherUser, contestId, &schemas.EditContest{
			Title: &newTitle,
		})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Invalid time range", func(t *testing.T) {
		cst := newContestServiceTest() // Create a new test instance
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		newEndTime := startTime.Add(-24 * time.Hour) // Before start
		err = cst.contestService.EditContest(cst.tx, currentUser, contestId, &schemas.EditContest{
			EndTime: &newEndTime,
		})
		assert.ErrorIs(t, err, errors.ErrInvalidTimeRange)
	})
}

func TestDeleteContest(t *testing.T) {
	cst := newContestServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		err = cst.contestService.DeleteContest(cst.tx, currentUser, contestId)
		assert.NoError(t, err)

		_, err = cst.contestService.GetContest(cst.tx, currentUser, contestId)
		assert.ErrorIs(t, err, errors.ErrContestNotFound)
	})

	t.Run("Contest not found", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		err := cst.contestService.DeleteContest(cst.tx, currentUser, 999)
		assert.ErrorIs(t, err, errors.ErrContestNotFound)
	})

	t.Run("Not authorized - teacher deleting another's contest", func(t *testing.T) {
		adminUser := cst.createUser(t, types.UserRoleAdmin)
		teacherUser := cst.createUser(t, types.UserRoleTeacher)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, adminUser, &schemas.CreateContest{
			Title:       "Admin Contest",
			Description: "Admin Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		err = cst.contestService.DeleteContest(cst.tx, teacherUser, contestId)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestAssignTasksToContest(t *testing.T) {
	cst := newContestServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		taskId := cst.createTask(t, currentUser, "Test Task")

		err = cst.contestService.AssignTasksToContest(cst.tx, currentUser, contestId, []int64{taskId})
		assert.NoError(t, err)

		contest, err := cst.contestService.GetContest(cst.tx, currentUser, contestId)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(contest.Tasks))
		assert.Equal(t, taskId, contest.Tasks[0].Id)
	})

	cst = newContestServiceTest()
	t.Run("Contest not found", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		taskId := cst.createTask(t, currentUser, "Test Task")

		err := cst.contestService.AssignTasksToContest(cst.tx, currentUser, 999, []int64{taskId})
		assert.ErrorIs(t, err, errors.ErrContestNotFound)
	})

	t.Run("Task not found", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		err = cst.contestService.AssignTasksToContest(cst.tx, currentUser, contestId, []int64{999})
		assert.ErrorIs(t, err, errors.ErrTaskNotFound)
	})
}

func TestUnAssignTasksFromContest(t *testing.T) {
	cst := newContestServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		taskId := cst.createTask(t, currentUser, "Test Task")

		err = cst.contestService.AssignTasksToContest(cst.tx, currentUser, contestId, []int64{taskId})
		assert.NoError(t, err)

		err = cst.contestService.UnAssignTasksFromContest(cst.tx, currentUser, contestId, []int64{taskId})
		assert.NoError(t, err)

		contest, err := cst.contestService.GetContest(cst.tx, currentUser, contestId)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(contest.Tasks))
	})
}

func TestGetContest(t *testing.T) {
	cst := newContestServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		contestId, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Test Contest",
			Description: "Test Description",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		contest, err := cst.contestService.GetContest(cst.tx, currentUser, contestId)
		assert.NoError(t, err)
		assert.Equal(t, "Test Contest", contest.Title)
		assert.Equal(t, "Test Description", contest.Description)
	})

	t.Run("Contest not found", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		_, err := cst.contestService.GetContest(cst.tx, currentUser, 999)
		assert.ErrorIs(t, err, errors.ErrContestNotFound)
	})
}

func TestGetAllContests(t *testing.T) {
	cst := newContestServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := cst.createUser(t, types.UserRoleAdmin)
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		
		_, err := cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Contest 1",
			Description: "Description 1",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		_, err = cst.contestService.Create(cst.tx, currentUser, &schemas.CreateContest{
			Title:       "Contest 2",
			Description: "Description 2",
			StartTime:   startTime,
			EndTime:     endTime,
		})
		assert.NoError(t, err)

		queryParams := map[string]interface{}{
			"limit":  uint64(10),
			"offset": uint64(0),
			"sort":   "",
		}
		contests, err := cst.contestService.GetAll(cst.tx, currentUser, queryParams)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(contests))
	})
}
