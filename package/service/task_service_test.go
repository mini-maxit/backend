package service_test

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
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type taskServiceTest struct {
	tx          *gorm.DB
	ur          repository.UserRepository
	tr          repository.TaskRepository
	gr          repository.GroupRepository
	taskService service.TaskService
	counter     int64
}

func newTaskServiceTest() *taskServiceTest {
	tx := &gorm.DB{}
	config := testutils.NewTestConfig()
	ur := testutils.NewMockUserRepository()
	gr := testutils.NewMockGroupRepository(ur)
	tr := testutils.NewMockTaskRepository(gr)
	io := testutils.NewMockInputOutputRepository()
	ts := service.NewTaskService(config.FileStorageURL, tr, io, ur, gr)

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
	userID, err := tst.ur.Create(tst.tx, &models.User{
		Name:         fmt.Sprintf("Test User %d", tst.counter),
		Surname:      fmt.Sprintf("Test Surname %d", tst.counter),
		Email:        fmt.Sprintf("email%d@email.com", tst.counter),
		Username:     fmt.Sprintf("testuser%d", tst.counter),
		Role:         role,
		PasswordHash: "password",
	})
	require.NoError(t, err)

	userModel, err := tst.ur.Get(tst.tx, userID)
	require.NoError(t, err)

	user := schemas.User{
		ID:   userModel.ID,
		Role: userModel.Role,
	}
	return user
}

func addDescription(t *testing.T, zipWriter *zip.Writer) {
	// Create description.pdf
	descriptionFile, err := zipWriter.Create("folder/description.pdf")
	require.NoError(t, err)
	_, err = descriptionFile.Write([]byte("This is a test description."))
	require.NoError(t, err)
}

func addInputOutputFiles(t *testing.T, zipWriter *zip.Writer, count int, inputDir, outputDir string) {
	for i := 1; i <= count; i++ {
		inputFile, err := zipWriter.Create(fmt.Sprintf("%s/%d.in", inputDir, i))
		require.NoError(t, err)
		_, err = inputFile.Write([]byte(fmt.Sprintf("Input data %d", i)))
		require.NoError(t, err)

		outputFile, err := zipWriter.Create(fmt.Sprintf("%s/%d.out", outputDir, i))
		require.NoError(t, err)
		_, err = outputFile.Write([]byte(fmt.Sprintf("Output data %d", i)))
		require.NoError(t, err)
	}
}

func (tst *taskServiceTest) createTestArchive(t *testing.T, caseType string) string {
	tempFile, err := os.CreateTemp(t.TempDir(), "test-archive-*.zip")
	require.NoError(t, err)
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
		require.NoError(t, err)
		_, err = outputFile.Write([]byte("Output data 3"))
		require.NoError(t, err)
	case "invalid_structure":
		addDescription(t, zipWriter)
		addInputOutputFiles(t, zipWriter, 4, "folder", "folder")
	case "single_file":
		// Create only one input and output file
		_, err := zipWriter.Create("1.in")
		require.NoError(t, err)
	case "nonexistent_file":
		// Create an invalid archive
		defer os.Remove(tempFile.Name())
	case "invalid_archive":
		defer os.Remove(tempFile.Name())
		file, err := os.CreateTemp(os.TempDir(), "test-archive-*.txt")
		require.NoError(t, err)
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
		require.NoError(t, err)
		addInputOutputFiles(t, zipWriter, 3, "folder/input", "folder/output")
		outputFile, err := zipWriter.Create(fmt.Sprintf("folder/output/%d.out", 4))
		require.NoError(t, err)
		_, err = outputFile.Write([]byte("Output data"))
		require.NoError(t, err)
	case "output_dir":
		// output dir contains another dir
		addDescription(t, zipWriter)
		_, err := zipWriter.Create("folder/output/another.out/output.out")
		require.NoError(t, err)
		addInputOutputFiles(t, zipWriter, 3, "folder/input", "folder/output")
		inputFile, err := zipWriter.Create("folder/input/another.in")
		require.NoError(t, err)
		_, err = inputFile.Write([]byte(fmt.Sprintf("Input data %d", 4)))
		require.NoError(t, err)
	}

	return tempFile.Name()
}

func TestCreateTask(t *testing.T) {
	tst := newTaskServiceTest()

	t.Run("Success", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		taskID, err := tst.taskService.Create(tst.tx, currentUser, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
	})

	// We want to have clear task repository for this state, and this is the quickest way
	tst = newTaskServiceTest()
	t.Run("Non unique title", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		taskID, err := tst.taskService.Create(tst.tx, currentUser, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: currentUser.ID,
		})
		require.NoError(t, err)
		assert.NotEqual(t, int64(0), taskID)
		taskID, err = tst.taskService.Create(tst.tx, currentUser, &schemas.Task{
			Title:     "Test Task",
			CreatedBy: currentUser.ID,
		})
		require.ErrorIs(t, err, errors.ErrTaskExists)
		assert.Equal(t, int64(0), taskID)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleStudent)
		taskID, err := tst.taskService.Create(tst.tx, currentUser, &schemas.Task{
			Title:     "Test Student Task",
			CreatedBy: currentUser.ID,
		})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Equal(t, int64(0), taskID)
	})
}

func TestGetTaskByTitle(t *testing.T) {
	tst := newTaskServiceTest()
	t.Run("Success", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: currentUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, currentUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		taskResp, err := tst.taskService.GetByTitle(tst.tx, task.Title)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		task, err := tst.taskService.GetByTitle(tst.tx, "Nonexistent Task")
		require.ErrorIs(t, err, errors.ErrTaskNotFound)
		assert.Nil(t, task)
	})
}

func TestGetAllTasks(t *testing.T) {
	tst := newTaskServiceTest()
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}
	t.Run("No tasks", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		tasks, err := tst.taskService.GetAll(tst.tx, currentUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: currentUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, currentUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		tasks, err := tst.taskService.GetAll(tst.tx, currentUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleStudent)
		tasks, err := tst.taskService.GetAll(tst.tx, currentUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})
}

func TestGetTask(t *testing.T) {
	tst := newTaskServiceTest()
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: adminUser.ID,
	}

	taskID, err := tst.taskService.Create(tst.tx, adminUser, task)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		taskResp, err := tst.taskService.Get(tst.tx, currentUser, taskID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Sucess with assigned task to user", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		err = tst.taskService.AssignToUsers(tst.tx, adminUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)
		taskResp, err := tst.taskService.Get(tst.tx, studentUser, taskID)
		t.Logf("Error: %v\n", err)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Sucess with assigned task to group", func(t *testing.T) {
		groupModel := &models.Group{
			Name: "Test Group",
		}
		groupID, err := tst.gr.Create(tst.tx, groupModel)
		require.NoError(t, err)
		err = tst.taskService.AssignToGroups(tst.tx, adminUser, taskID, []int64{groupID})
		require.NoError(t, err)
		studentUser := tst.createUser(t, types.UserRoleStudent)
		err = tst.gr.AddUser(tst.tx, groupID, studentUser.ID)
		require.NoError(t, err)
		taskResp, err := tst.taskService.Get(tst.tx, studentUser, taskID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Success with created task", func(t *testing.T) {
		teacherUser := tst.createUser(t, types.UserRoleTeacher)
		task := &schemas.Task{
			Title:     "Test Task 2",
			CreatedBy: teacherUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, teacherUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		taskResp, err := tst.taskService.Get(tst.tx, teacherUser, taskID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Not authorized", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		taskResp, err := tst.taskService.Get(tst.tx, studentUser, taskID)
		require.NoError(t, err)
		assert.NotEmpty(t, taskResp)
	})
}

func TestAssignTaskToUsers(t *testing.T) {
	tst := newTaskServiceTest()
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: adminUser.ID,
	}

	taskID, err := tst.taskService.Create(tst.tx, adminUser, task)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		err := tst.taskService.AssignToUsers(tst.tx, adminUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		err := tst.taskService.AssignToUsers(tst.tx, adminUser, 0, []int64{studentUser.ID})
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		err := tst.taskService.AssignToUsers(tst.tx, studentUser, taskID, []int64{studentUser.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		teacherUser := tst.createUser(t, types.UserRoleTeacher)
		err := tst.taskService.AssignToUsers(tst.tx, teacherUser, taskID, []int64{teacherUser.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestAssignTaskToGroups(t *testing.T) {
	tst := newTaskServiceTest()
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: adminUser.ID,
	}

	taskID, err := tst.taskService.Create(tst.tx, adminUser, task)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		groupModel := &models.Group{
			Name: "Test Group",
		}
		groupID, err := tst.gr.Create(tst.tx, groupModel)
		require.NoError(t, err)
		err = tst.taskService.AssignToGroups(tst.tx, adminUser, taskID, []int64{groupID})
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		groupModel := &models.Group{
			Name: "Test Group",
		}
		groupID, err := tst.gr.Create(tst.tx, groupModel)
		require.NoError(t, err)
		err = tst.taskService.AssignToGroups(tst.tx, adminUser, 0, []int64{groupID})
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized student", func(t *testing.T) {
		groupModel := &models.Group{
			Name: "Test Group",
		}
		groupID, err := tst.gr.Create(tst.tx, groupModel)
		require.NoError(t, err)
		studentUser := tst.createUser(t, types.UserRoleStudent)
		err = tst.gr.AddUser(tst.tx, groupID, studentUser.ID)
		require.NoError(t, err)
		err = tst.taskService.AssignToGroups(tst.tx, studentUser, taskID, []int64{groupID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		teacherUser := tst.createUser(t, types.UserRoleTeacher)
		groupModel := &models.Group{
			Name:      "Test Group",
			CreatedBy: teacherUser.ID + 1,
		}
		groupID, err := tst.gr.Create(tst.tx, groupModel)
		require.NoError(t, err)
		err = tst.taskService.AssignToGroups(tst.tx, teacherUser, taskID, []int64{groupID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetAllAssignedTasks(t *testing.T) {
	tst := newTaskServiceTest()
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: adminUser.ID,
	}

	taskID, err := tst.taskService.Create(tst.tx, adminUser, task)
	require.NoError(t, err)

	t.Run("No tasks", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		tasks, err := tst.taskService.GetAllAssigned(tst.tx, studentUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		err := tst.taskService.AssignToUsers(tst.tx, adminUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)
		group := &models.Group{
			Name: "Test Group",
		}
		groupID, err := tst.gr.Create(tst.tx, group)
		require.NoError(t, err)
		err = tst.gr.AddUser(tst.tx, groupID, studentUser.ID)
		require.NoError(t, err)
		err = tst.taskService.AssignToGroups(tst.tx, adminUser, taskID, []int64{groupID})
		require.NoError(t, err)
		tasks, err := tst.taskService.GetAllAssigned(tst.tx, studentUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 2)
	})
}

func TestDeleteTask(t *testing.T) {
	tst := newTaskServiceTest()
	t.Run("Success", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: currentUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, currentUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		err = tst.taskService.Delete(tst.tx, currentUser, taskID)
		require.NoError(t, err)
		_, err = tst.taskService.Get(tst.tx, currentUser, taskID)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		err := tst.taskService.Delete(tst.tx, currentUser, 0)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleStudent)
		adminUser := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: currentUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, adminUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		err = tst.taskService.Delete(tst.tx, currentUser, taskID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}
func TestUpdateTask(t *testing.T) {
	tst := newTaskServiceTest()
	adminUser := tst.createUser(t, types.UserRoleAdmin)

	t.Run("Success", func(t *testing.T) {
		currentUser := tst.createUser(t, types.UserRoleAdmin)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: currentUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, currentUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		newTitle := "Updated Task"
		updatedTask := &schemas.EditTask{Title: &newTitle}
		err = tst.taskService.Edit(tst.tx, adminUser, taskID, updatedTask)
		require.NoError(t, err)
		taskResp, err := tst.taskService.Get(tst.tx, currentUser, taskID)
		require.NoError(t, err)
		assert.Equal(t, *updatedTask.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})
	t.Run("Nonexistent task", func(t *testing.T) {
		newTitle := "Updated Task"
		updatedTask := &schemas.EditTask{Title: &newTitle}
		err := tst.taskService.Edit(tst.tx, adminUser, 0, updatedTask)
		require.ErrorIs(t, err, errors.ErrTaskNotFound)
	})
}

func TestTaskGetAllForGroup(t *testing.T) {
	tst := newTaskServiceTest()
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	group := &models.Group{
		Name: "Test Group",
	}
	groupID, err := tst.gr.Create(tst.tx, group)
	require.NoError(t, err)

	t.Run("No tasks", func(t *testing.T) {
		tasks, err := tst.taskService.GetAllForGroup(tst.tx, adminUser, groupID, queryParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, adminUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		task = &schemas.Task{
			Title:     "Test Task2",
			CreatedBy: adminUser.ID,
		}
		taskID, err = tst.taskService.Create(tst.tx, adminUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		err = tst.taskService.AssignToGroups(tst.tx, adminUser, taskID, []int64{groupID})
		require.NoError(t, err)
		tasks, err := tst.taskService.GetAllForGroup(tst.tx, adminUser, groupID, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 1)
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Not authorized", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		tasks, err := tst.taskService.GetAllForGroup(tst.tx, studentUser, groupID, queryParams)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Empty(t, tasks)
	})
}

func TestGetAllCreatedTasks(t *testing.T) {
	tst := newTaskServiceTest()
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	teacherUser := tst.createUser(t, types.UserRoleTeacher)

	t.Run("No tasks", func(t *testing.T) {
		tasks, err := tst.taskService.GetAllCreated(tst.tx, adminUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success with admin", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, adminUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		tasks, err := tst.taskService.GetAllCreated(tst.tx, adminUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 1)
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Teacher Task",
			CreatedBy: teacherUser.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, teacherUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		tasks, err := tst.taskService.GetAllCreated(tst.tx, teacherUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 1)
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Different teachers", func(t *testing.T) {
		teacherUser2 := tst.createUser(t, types.UserRoleTeacher)
		task := &schemas.Task{
			Title:     "Teacher Task 2",
			CreatedBy: teacherUser2.ID,
		}
		taskID, err := tst.taskService.Create(tst.tx, teacherUser2, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
		tasks, err := tst.taskService.GetAllCreated(tst.tx, teacherUser2, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 1)
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Not authorized", func(t *testing.T) {
		studentUser := tst.createUser(t, types.UserRoleStudent)
		tasks, err := tst.taskService.GetAllCreated(tst.tx, studentUser, queryParams)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Empty(t, tasks)
	})
}

func TestUnAssignTaskFromUsers(t *testing.T) {
	tst := newTaskServiceTest()
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	teacherUser := tst.createUser(t, types.UserRoleTeacher)
	studentUser := tst.createUser(t, types.UserRoleStudent)

	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: teacherUser.ID,
	}
	taskID, err := tst.taskService.Create(tst.tx, teacherUser, task)
	require.NoError(t, err)
	assert.NotEqual(t, 0, taskID)

	t.Run("Success with admin", func(t *testing.T) {
		err := tst.taskService.AssignToUsers(tst.tx, adminUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)

		err = tst.taskService.UnassignFromUsers(tst.tx, adminUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		err := tst.taskService.AssignToUsers(tst.tx, teacherUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)

		err = tst.taskService.UnassignFromUsers(tst.tx, teacherUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := tst.taskService.AssignToUsers(tst.tx, teacherUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)

		err = tst.taskService.UnassignFromUsers(tst.tx, studentUser, taskID, []int64{studentUser.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestUnAssignTaskFromGroups(t *testing.T) {
	tst := newTaskServiceTest()
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	teacherUser := tst.createUser(t, types.UserRoleTeacher)
	studentUser := tst.createUser(t, types.UserRoleStudent)

	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: teacherUser.ID,
	}
	taskID, err := tst.taskService.Create(tst.tx, teacherUser, task)
	require.NoError(t, err)
	assert.NotEqual(t, 0, taskID)

	group := &models.Group{
		Name: "Test Group",
	}
	groupID, err := tst.gr.Create(tst.tx, group)
	require.NoError(t, err)

	t.Run("Success with admin", func(t *testing.T) {
		err := tst.taskService.AssignToGroups(tst.tx, adminUser, taskID, []int64{groupID})
		require.NoError(t, err)

		err = tst.taskService.UnassignFromGroups(tst.tx, adminUser, taskID, []int64{groupID})
		require.NoError(t, err)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		err := tst.taskService.AssignToGroups(tst.tx, teacherUser, taskID, []int64{groupID})
		require.NoError(t, err)

		err = tst.taskService.UnassignFromGroups(tst.tx, teacherUser, taskID, []int64{groupID})
		require.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := tst.taskService.AssignToGroups(tst.tx, teacherUser, taskID, []int64{groupID})
		require.NoError(t, err)

		err = tst.taskService.UnassignFromGroups(tst.tx, studentUser, taskID, []int64{groupID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestCreateInputOutput(t *testing.T) {
	tst := newTaskServiceTest()
	adminUser := tst.createUser(t, types.UserRoleAdmin)
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: adminUser.ID,
	}

	taskID, err := tst.taskService.Create(tst.tx, adminUser, task)
	require.NoError(t, err)
	assert.NotEqual(t, 0, taskID)

	t.Run("Success", func(t *testing.T) {
		pathToArchive := tst.createTestArchive(t, "valid")
		defer os.Remove(pathToArchive)
		err := tst.taskService.CreateInputOutput(tst.tx, taskID, pathToArchive)
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		pathToArchive := tst.createTestArchive(t, "valid")
		defer os.Remove(pathToArchive)
		err := tst.taskService.CreateInputOutput(tst.tx, -1, pathToArchive)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Invalid archive path", func(t *testing.T) {
		err := tst.taskService.CreateInputOutput(tst.tx, taskID, "INVALIDPATH")
		require.Error(t, err)
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
		t.Run(tt.name, func(t *testing.T) {
			pathToArchive := tst.createTestArchive(t, tt.caseType)
			numFiles, err := tst.taskService.ParseInputOutput(pathToArchive)
			if tt.isError {
				if tt.expectedError != nil {
					require.ErrorIs(t, err, tt.expectedError)
				} else {
					require.Error(t, err)
				}
				assert.Equal(t, tt.expected, numFiles)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, numFiles)
			}
		})
	}
}
