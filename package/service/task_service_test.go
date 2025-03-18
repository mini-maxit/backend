package service

import (
	"archive/zip"
	"fmt"
	"os"
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
	io := testutils.NewMockInputOutputRepository()
	ts := NewTaskService(config.FileStorageUrl, tr, io, ur, gr)

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

func addDescription(t *testing.T, zipWriter *zip.Writer) {
	// Create description.pdf
	descriptionFile, err := zipWriter.Create("folder/description.pdf")
	assert.NoError(t, err)
	_, err = descriptionFile.Write([]byte("This is a test description."))
	assert.NoError(t, err)
}

func addInputOutputFiles(t *testing.T, zipWriter *zip.Writer, count int, inputDir, outputDir string) {
	for i := 1; i <= count; i++ {
		inputFile, err := zipWriter.Create(fmt.Sprintf("%s/%d.in", inputDir, i))
		assert.NoError(t, err)
		_, err = inputFile.Write([]byte(fmt.Sprintf("Input data %d", i)))
		assert.NoError(t, err)

		outputFile, err := zipWriter.Create(fmt.Sprintf("%s/%d.out", outputDir, i))
		assert.NoError(t, err)
		_, err = outputFile.Write([]byte(fmt.Sprintf("Output data %d", i)))
		assert.NoError(t, err)
	}
}

func (tst *taskServiceTest) createTestArchive(t *testing.T, caseType string) string {
	tempFile, err := os.CreateTemp(os.TempDir(), "test-archive-*.zip")
	assert.NoError(t, err)
	defer tempFile.Close()

	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// Create input and output files based on caseType
	switch caseType {
	case "valid":
		addDescription(t, zipWriter)
		addInputOutputFiles(t, zipWriter, 4, "folder/input", "folder/output")
	case "missing_files":
		addDescription(t, zipWriter)
		addInputOutputFiles(t, zipWriter, 2, "folder/input", "folder/output")
		outputFile, err := zipWriter.Create("folder/output/3.out")
		assert.NoError(t, err)
		_, err = outputFile.Write([]byte("Output data 3"))
		assert.NoError(t, err)
	case "invalid_structure":
		addDescription(t, zipWriter)
		addInputOutputFiles(t, zipWriter, 4, "folder", "folder")
	case "single_file":
		// Create only one input and output file
		_, err := zipWriter.Create("1.in")
		assert.NoError(t, err)
	case "nonexistent_file":
		// Create an invalid archive
		defer os.Remove(tempFile.Name())
	case "invalid_archive":
		defer os.Remove(tempFile.Name())
		file, err := os.CreateTemp(os.TempDir(), "test-archive-*.txt")
		assert.NoError(t, err)
		return file.Name()
	case "no_output":
		// Create only input files
		addDescription(t, zipWriter)
		addInputOutputFiles(t, zipWriter, 4, "folder/input", "folder/invalid")
	case "no_input":
		// Create only output files
		addDescription(t, zipWriter)
		addInputOutputFiles(t, zipWriter, 4, "folder/invalid", "folder/output")
	case "input_dir":
		// input dir contains another dir
		addDescription(t, zipWriter)
		_, err := zipWriter.Create("folder/input/another.in/input.in")
		assert.NoError(t, err)
		addInputOutputFiles(t, zipWriter, 3, "folder/input", "folder/output")
		outputFile, err := zipWriter.Create(fmt.Sprintf("folder/output/%d.out", 4))
		assert.NoError(t, err)
		_, err = outputFile.Write([]byte("Output data"))
		assert.NoError(t, err)
	case "output_dir":
		// output dir contains another dir
		addDescription(t, zipWriter)
		_, err := zipWriter.Create("folder/output/another.out/output.out")
		assert.NoError(t, err)
		addInputOutputFiles(t, zipWriter, 3, "folder/input", "folder/output")
		inputFile, err := zipWriter.Create("folder/input/another.in")
		assert.NoError(t, err)
		_, err = inputFile.Write([]byte(fmt.Sprintf("Input data %d", 4)))
		assert.NoError(t, err)
	}

	return tempFile.Name()
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

	t.Run("Not authorized teacher", func(t *testing.T) {
		teacher_user := tst.createUser(t, types.UserRoleTeacher)
		err := tst.taskService.AssignTaskToUsers(tst.tx, teacher_user, taskId, []int64{teacher_user.Id})
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

	t.Run("Not authorized student", func(t *testing.T) {
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

	t.Run("Not authorized teacher", func(t *testing.T) {
		teacher_user := tst.createUser(t, types.UserRoleTeacher)
		group_model := &models.Group{
			Name:      "Test Group",
			CreatedBy: teacher_user.Id + 1,
		}
		groupId, err := tst.gr.CreateGroup(tst.tx, group_model)
		assert.NoError(t, err)
		err = tst.taskService.AssignTaskToGroups(tst.tx, teacher_user, taskId, []int64{groupId})
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

func TestGetAllForGroup(t *testing.T) {
	tst := newTaskServiceTest()
	queryParams := map[string]interface{}{"limit": uint64(10), "offset": uint64(0), "sort": "id:asc"}
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	group := &models.Group{
		Name: "Test Group",
	}
	groupId, err := tst.gr.CreateGroup(tst.tx, group)
	assert.NoError(t, err)

	t.Run("No tasks", func(t *testing.T) {
		tasks, err := tst.taskService.GetAllForGroup(tst.tx, admin_user, groupId, queryParams)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: admin_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, admin_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		task = &schemas.Task{
			Title:     "Test Task2",
			CreatedBy: admin_user.Id,
		}
		taskId, err = tst.taskService.Create(tst.tx, admin_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		err = tst.taskService.AssignTaskToGroups(tst.tx, admin_user, taskId, []int64{groupId})
		assert.NoError(t, err)
		tasks, err := tst.taskService.GetAllForGroup(tst.tx, admin_user, groupId, queryParams)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, 1, len(tasks))
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Not authorized", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		tasks, err := tst.taskService.GetAllForGroup(tst.tx, student_user, groupId, queryParams)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Empty(t, tasks)
	})
}

func TestGetAllCreatedTasks(t *testing.T) {
	tst := newTaskServiceTest()
	queryParams := map[string]interface{}{"limit": uint64(10), "offset": uint64(0), "sort": "id:asc"}
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	teacher_user := tst.createUser(t, types.UserRoleTeacher)

	t.Run("No tasks", func(t *testing.T) {
		tasks, err := tst.taskService.GetAllCreatedTasks(tst.tx, admin_user, queryParams)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success with admin", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: admin_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, admin_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		tasks, err := tst.taskService.GetAllCreatedTasks(tst.tx, admin_user, queryParams)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, 1, len(tasks))
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Teacher Task",
			CreatedBy: teacher_user.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, teacher_user, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		tasks, err := tst.taskService.GetAllCreatedTasks(tst.tx, teacher_user, queryParams)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, 1, len(tasks))
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Different teachers", func(t *testing.T) {
		teacher_user2 := tst.createUser(t, types.UserRoleTeacher)
		task := &schemas.Task{
			Title:     "Teacher Task 2",
			CreatedBy: teacher_user2.Id,
		}
		taskId, err := tst.taskService.Create(tst.tx, teacher_user2, task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, taskId)
		tasks, err := tst.taskService.GetAllCreatedTasks(tst.tx, teacher_user2, queryParams)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, 1, len(tasks))
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Not authorized", func(t *testing.T) {
		student_user := tst.createUser(t, types.UserRoleStudent)
		tasks, err := tst.taskService.GetAllCreatedTasks(tst.tx, student_user, queryParams)
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Empty(t, tasks)
	})
}

func TestUnAssignTaskFromUsers(t *testing.T) {
	tst := newTaskServiceTest()
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	teacher_user := tst.createUser(t, types.UserRoleTeacher)
	student_user := tst.createUser(t, types.UserRoleStudent)

	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: teacher_user.Id,
	}
	taskId, err := tst.taskService.Create(tst.tx, teacher_user, task)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, taskId)

	t.Run("Success with admin", func(t *testing.T) {
		err := tst.taskService.AssignTaskToUsers(tst.tx, admin_user, taskId, []int64{student_user.Id})
		assert.NoError(t, err)

		err = tst.taskService.UnAssignTaskFromUsers(tst.tx, admin_user, taskId, []int64{student_user.Id})
		assert.NoError(t, err)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		err := tst.taskService.AssignTaskToUsers(tst.tx, teacher_user, taskId, []int64{student_user.Id})
		assert.NoError(t, err)

		err = tst.taskService.UnAssignTaskFromUsers(tst.tx, teacher_user, taskId, []int64{student_user.Id})
		assert.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := tst.taskService.AssignTaskToUsers(tst.tx, teacher_user, taskId, []int64{student_user.Id})
		assert.NoError(t, err)

		err = tst.taskService.UnAssignTaskFromUsers(tst.tx, student_user, taskId, []int64{student_user.Id})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestUnAssignTaskFromGroups(t *testing.T) {
	tst := newTaskServiceTest()
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	teacher_user := tst.createUser(t, types.UserRoleTeacher)
	student_user := tst.createUser(t, types.UserRoleStudent)

	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: teacher_user.Id,
	}
	taskId, err := tst.taskService.Create(tst.tx, teacher_user, task)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, taskId)

	group := &models.Group{
		Name: "Test Group",
	}
	groupId, err := tst.gr.CreateGroup(tst.tx, group)
	assert.NoError(t, err)

	t.Run("Success with admin", func(t *testing.T) {
		err := tst.taskService.AssignTaskToGroups(tst.tx, admin_user, taskId, []int64{groupId})
		assert.NoError(t, err)

		err = tst.taskService.UnAssignTaskFromGroups(tst.tx, admin_user, taskId, []int64{groupId})
		assert.NoError(t, err)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		err := tst.taskService.AssignTaskToGroups(tst.tx, teacher_user, taskId, []int64{groupId})
		assert.NoError(t, err)

		err = tst.taskService.UnAssignTaskFromGroups(tst.tx, teacher_user, taskId, []int64{groupId})
		assert.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := tst.taskService.AssignTaskToGroups(tst.tx, teacher_user, taskId, []int64{groupId})
		assert.NoError(t, err)

		err = tst.taskService.UnAssignTaskFromGroups(tst.tx, student_user, taskId, []int64{groupId})
		assert.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestCreateInputOutput(t *testing.T) {
	tst := newTaskServiceTest()
	admin_user := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: admin_user.Id,
	}

	taskId, err := tst.taskService.Create(tst.tx, admin_user, task)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, taskId)

	t.Run("Success", func(t *testing.T) {
		pathToArchive := tst.createTestArchive(t, "valid")
		defer os.Remove(pathToArchive)
		err := tst.taskService.CreateInputOutput(tst.tx, taskId, pathToArchive)
		assert.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		pathToArchive := tst.createTestArchive(t, "valid")
		defer os.Remove(pathToArchive)
		err := tst.taskService.CreateInputOutput(tst.tx, -1, pathToArchive)
		assert.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Invalid archive path", func(t *testing.T) {
		err := tst.taskService.CreateInputOutput(tst.tx, taskId, "INVALIDPATH")
		assert.Error(t, err)
	})
}

func TestParseInputOutput(t *testing.T) {
	tst := newTaskServiceTest()
	tests := []struct {
		name          string
		caseType      string
		expected      int
		isError       bool
		expectedError error
	}{{
		name:     "Valid archive",
		caseType: "valid",
		expected: 4,
	}, {
		name:     "Missing files",
		caseType: "missing_files",
		expected: -1,
		isError:  true,
	}, {
		name:     "Invalid structure",
		caseType: "invalid_structure",
		expected: -1,
		isError:  true,
	}, {
		name:     "Single file",
		caseType: "single_file",
		expected: -1,
		isError:  true,
	}, {
		name:          "Nonexistent file",
		caseType:      "nonexistent_file",
		expected:      -1,
		isError:       true,
		expectedError: errors.ErrFileOpen,
	}, {
		name:          "Invalid archive",
		caseType:      "invalid_archive",
		expected:      -1,
		isError:       true,
		expectedError: errors.ErrDecompressArchive,
	}, {
		name:          "No output dir",
		caseType:      "no_output",
		expected:      -1,
		isError:       true,
		expectedError: errors.ErrNoOutputDirectory,
	}, {
		name:          "No input dir",
		caseType:      "no_input",
		expected:      -1,
		isError:       true,
		expectedError: errors.ErrNoInputDirectory,
	}, {
		name:          "Input contains directories",
		caseType:      "input_dir",
		expected:      -1,
		isError:       true,
		expectedError: errors.ErrInputContainsDir,
	},
	}
	for _, tt := range tests {
		pathToArchive := tst.createTestArchive(t, tt.caseType)
		numFiles, err := tst.taskService.ParseInputOutput(pathToArchive)
		if tt.isError {
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.expected, numFiles)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, numFiles)
		}
	}
}
