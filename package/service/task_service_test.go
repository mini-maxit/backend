package service_test

import (
	"archive/zip"
	"fmt"
	"gorm.io/gorm"
	"os"
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/filestorage"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
		_, err = fmt.Fprintf(inputFile, "Input data %d", i)
		require.NoError(t, err)

		outputFile, err := zipWriter.Create(fmt.Sprintf("%s/%d.out", outputDir, i))
		require.NoError(t, err)
		_, err = fmt.Fprintf(outputFile, "Output data %d", i)
		require.NoError(t, err)
	}
}

func createTestArchive(t *testing.T, caseType string) string {
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
		file, err := os.CreateTemp(t.TempDir(), "test-archive-*.txt")
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
		_, err = fmt.Fprintf(inputFile, "Input data %d", 4)
		require.NoError(t, err)
	}

	return tempFile.Name()
}

func TestCreateTask(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}

	t.Run("Success", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		ur.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&models.User{ID: 1, Role: types.UserRoleAdmin}, nil).Times(1)
		tr.EXPECT().GetByTitle(tx, task.Title).Return(nil, gorm.ErrRecordNotFound).Times(1)
		tr.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(1), nil).Times(1)

		taskID, err := ts.Create(tx, adminUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
	})

	t.Run("Non unique title", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetByTitle(tx, task.Title).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		taskID, err := ts.Create(tx, adminUser, task)
		require.ErrorIs(t, err, errors.ErrTaskExists)
		assert.Equal(t, int64(0), taskID)
	})

	t.Run("Not authorized", func(t *testing.T) {
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		taskID, err := ts.Create(tx, studentUser, &schemas.Task{
			Title:     "Test Student Task",
			CreatedBy: studentUser.ID,
		})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Equal(t, int64(0), taskID)
	})
}

func TestGetTaskByTitle(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}

	t.Run("Success", func(t *testing.T) {
		taskID := int64(1)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetByTitle(tx, task.Title).Return(&models.Task{
			ID:        taskID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		taskResp, err := ts.GetByTitle(tx, task.Title)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		taskTitle := "Nonexistent Task"
		tr.EXPECT().GetByTitle(tx, taskTitle).Return(nil, errors.ErrTaskNotFound).Times(1)
		task, err := ts.GetByTitle(tx, taskTitle)
		require.ErrorIs(t, err, errors.ErrTaskNotFound)
		assert.Nil(t, task)
	})
}

func TestGetAllTasks(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)

	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}

	t.Run("No tasks", func(t *testing.T) {
		tr.EXPECT().GetAll(tx,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.Task{}, nil).Times(1)
		tasks, err := ts.GetAll(tx, adminUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		currentUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
		tasks := []models.Task{
			{
				ID:        1,
				Title:     "Test Task",
				CreatedBy: currentUser.ID,
			},
		}
		tr.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tasks, nil).Times(1)

		resultTasks, err := ts.GetAll(tx, currentUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, resultTasks)
	})
}

func TestGetTask(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	config := testutils.NewTestConfig()
	fs, err := filestorage.NewFileStorageService(config.FileStorageURL)
	require.NoError(t, err)
	ts := service.NewTaskService(fs, fr, tr, io, ur, gr)

	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: adminUser.ID,
	}

	t.Run("Success", func(t *testing.T) {
		tr.EXPECT().Get(tx, task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		taskResp, err := ts.Get(tx, adminUser, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Sucess with assigned task to user", func(t *testing.T) {
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		tr.EXPECT().Get(gomock.Any(), task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)

		// tr.EXPECT().IsAssignedToUser(tx, task.ID, studentUser.ID).Return(true, nil).Times(1)
		taskResp, err := ts.Get(tx, studentUser, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, taskResp.ID)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Sucess with assigned task to group", func(t *testing.T) {
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}

		tr.EXPECT().Get(tx, task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)

		taskResp, err := ts.Get(tx, studentUser, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, taskResp.ID)
		assert.Equal(t, task.Title, taskResp.Title)
	})

	t.Run("Success with created task", func(t *testing.T) {
		teacherUser := schemas.User{ID: 2, Role: types.UserRoleTeacher}
		tr.EXPECT().Get(tx, task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		taskResp, err := ts.Get(tx, teacherUser, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, taskResp.ID)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Fail with non existent task", func(t *testing.T) {
		taskID := int64(0)
		tr.EXPECT().Get(tx, taskID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		taskResp, err := ts.Get(tx, adminUser, taskID)
		require.ErrorIs(t, err, errors.ErrNotFound)
		assert.Nil(t, taskResp)
	})
}

func TestAssignTaskToUsers(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)

	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	taskID := int64(1)

	t.Run("Success", func(t *testing.T) {
		studentUser := schemas.User{ID: 2}
		ur.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&models.User{ID: 2, Role: types.UserRoleStudent}, nil).Times(1)
		tr.EXPECT().Get(tx, taskID).Return(&models.Task{
			ID: taskID,
		}, nil).Times(1)
		tr.EXPECT().IsAssignedToUser(tx, taskID, studentUser.ID).Return(false, nil).Times(1)
		tr.EXPECT().AssignToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		err := ts.AssignToUsers(tx, adminUser, taskID, []int64{studentUser.ID})
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		studentUser := schemas.User{ID: 2}
		tr.EXPECT().Get(tx, taskID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := ts.AssignToUsers(tx, adminUser, taskID, []int64{studentUser.ID})
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		err := ts.AssignToUsers(tx, studentUser, taskID, []int64{studentUser.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		teacherUser := schemas.User{ID: 2, Role: types.UserRoleTeacher}
		tr.EXPECT().Get(tx, taskID).Return(&models.Task{
			ID:        taskID,
			CreatedBy: adminUser.ID,
		}, nil).Times(1)
		err := ts.AssignToUsers(tx, teacherUser, taskID, []int64{teacherUser.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestAssignTaskToGroups(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)

	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	task := &schemas.Task{
		ID:        int64(1),
		Title:     "Test Task",
		CreatedBy: adminUser.ID,
	}

	t.Run("Success", func(t *testing.T) {
		groupID := int64(1)
		gr.EXPECT().Get(tx, groupID).Return(&models.Group{
			ID:   groupID,
			Name: "Test Group",
		}, nil).Times(1)
		tr.EXPECT().Get(tx, task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		tr.EXPECT().AssignToGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		tr.EXPECT().IsAssignedToGroup(tx, task.ID, groupID).Return(false, nil).Times(1)
		err := ts.AssignToGroups(tx, adminUser, task.ID, []int64{groupID})
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		groupID := int64(1)
		tr.EXPECT().Get(tx, int64(0)).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := ts.AssignToGroups(tx, adminUser, int64(0), []int64{groupID})
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized student", func(t *testing.T) {
		groupID := int64(1)
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		err := ts.AssignToGroups(tx, studentUser, task.ID, []int64{groupID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})

	t.Run("Not authorized teacher", func(t *testing.T) {
		teacherUser := schemas.User{ID: 2, Role: types.UserRoleTeacher}
		groupID := int64(1)
		tr.EXPECT().Get(tx, task.ID).Return(&models.Task{
			ID: task.ID,
		}, nil).Times(1)
		err := ts.AssignToGroups(tx, teacherUser, task.ID, []int64{groupID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestGetAllAssignedTasks(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}

	t.Run("No tasks", func(t *testing.T) {
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		tr.EXPECT().GetAllAssigned(
			tx,
			studentUser.ID,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.Task{}, nil).Times(1)

		tasks, err := ts.GetAllAssigned(tx, studentUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		taskID := int64(1)

		tr.EXPECT().GetAllAssigned(
			tx,
			studentUser.ID,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.Task{
			{ID: taskID, Title: "Test Task 1", CreatedBy: adminUser.ID},
			{ID: taskID, Title: "Test Task 2", CreatedBy: adminUser.ID},
		}, nil).Times(1)

		tasks, err := ts.GetAllAssigned(tx, studentUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 2)
	})
}

func TestDeleteTask(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	taskID := int64(1)

	t.Run("Success for admin", func(t *testing.T) {
		tr.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		tr.EXPECT().Get(gomock.Any(), taskID).Return(&models.Task{ID: taskID}, nil).Times(1)
		err := ts.Delete(tx, adminUser, taskID)
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		tr.EXPECT().Get(gomock.Any(), int64(0)).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := ts.Delete(tx, adminUser, 0)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		currentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		err := ts.Delete(tx, currentUser, taskID)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestEditTask(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	taskID := int64(1)

	t.Run("Success", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().Get(tx, taskID).Return(&models.Task{
			ID:        taskID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		tr.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		newTitle := "Updated Task"
		updatedTask := &schemas.EditTask{Title: &newTitle}
		err := ts.Edit(tx, adminUser, taskID, updatedTask)
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		newTitle := "Updated Task"
		updatedTask := &schemas.EditTask{Title: &newTitle}
		tr.EXPECT().Get(tx, int64(0)).Return(nil, errors.ErrTaskNotFound).Times(1)
		ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
		err := ts.Edit(tx, adminUser, 0, updatedTask)
		require.ErrorIs(t, err, errors.ErrTaskNotFound)
	})
}

func TestTaskGetAllForGroup(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	taskID := int64(1)
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}
	group := &models.Group{
		ID:   int64(1),
		Name: "Test Group",
	}

	t.Run("No tasks", func(t *testing.T) {
		groupID := int64(1)
		tr.EXPECT().GetAllForGroup(
			tx,
			groupID,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.Task{}, nil).Times(1)
		tasks, err := ts.GetAllForGroup(tx, adminUser, groupID, queryParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		tr.EXPECT().GetAllForGroup(
			tx,
			group.ID,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.Task{
			{
				ID: taskID,
			},
		}, nil).Times(1)
		tasks, err := ts.GetAllForGroup(tx, adminUser, group.ID, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 1)
		assert.Equal(t, tasks[0].ID, taskID)
	})

	t.Run("Not authorized", func(t *testing.T) {
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		tasks, err := ts.GetAllForGroup(tx, studentUser, group.ID, queryParams)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Empty(t, tasks)
	})
}

func TestGetAllCreatedTasks(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	taskID := int64(1)
	queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "id:asc"}

	t.Run("No tasks", func(t *testing.T) {
		tr.EXPECT().GetAllCreated(
			tx,
			adminUser.ID,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.Task{}, nil).Times(1)
		tasks, err := ts.GetAllCreated(tx, adminUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success with admin", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetAllCreated(
			tx,
			adminUser.ID,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.Task{
			{
				ID:        taskID,
				Title:     task.Title,
				CreatedBy: task.CreatedBy,
			},
		}, nil).Times(1)
		tasks, err := ts.GetAllCreated(tx, adminUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 1)
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		teacherUser := schemas.User{ID: 2, Role: types.UserRoleTeacher}
		task := &schemas.Task{
			Title:     "Teacher Task",
			CreatedBy: teacherUser.ID,
		}
		tr.EXPECT().GetAllCreated(
			tx,
			teacherUser.ID,
			queryParams["limit"],
			queryParams["offset"],
			queryParams["sort"],
		).Return([]models.Task{
			{
				ID:        taskID,
				Title:     task.Title,
				CreatedBy: task.CreatedBy,
			},
		}, nil).Times(1)
		tasks, err := ts.GetAllCreated(tx, teacherUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 1)
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Different teachers", func(t *testing.T) {
		teacherUser := schemas.User{ID: 2, Role: types.UserRoleTeacher}
		task := &schemas.Task{
			Title:     "Teacher Task 2",
			CreatedBy: teacherUser.ID,
		}
		tr.EXPECT().GetAllCreated(
			tx,
			teacherUser.ID,
			queryParams["limit"],
			queryParams["offset"], queryParams["sort"],
		).Return([]models.Task{
			{
				ID:        taskID,
				Title:     task.Title,
				CreatedBy: task.CreatedBy,
			},
		}, nil).Times(1)
		tasks, err := ts.GetAllCreated(tx, teacherUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Len(t, tasks, 1)
		assert.Equal(t, task.Title, tasks[0].Title)
		assert.Equal(t, task.CreatedBy, tasks[0].CreatedBy)
	})

	t.Run("Not authorized", func(t *testing.T) {
		studentUser := schemas.User{ID: 2, Role: types.UserRoleStudent}
		tasks, err := ts.GetAllCreated(tx, studentUser, queryParams)
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
		assert.Empty(t, tasks)
	})
}

func TestUnAssignTaskFromUsers(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	teacherUser := schemas.User{ID: 2, Role: types.UserRoleTeacher}
	studentUser := schemas.User{ID: 3, Role: types.UserRoleStudent}
	task := &models.Task{
		ID:        int64(1),
		CreatedBy: teacherUser.ID,
	}

	t.Run("Success with admin", func(t *testing.T) {
		tr.EXPECT().IsAssignedToUser(tx, task.ID, studentUser.ID).Return(true, nil).Times(1)
		tr.EXPECT().UnassignFromUser(tx, task.ID, studentUser.ID).Return(nil).Times(1)
		tr.EXPECT().Get(tx, task.ID).Return(task, nil).Times(1)

		err := ts.UnassignFromUsers(tx, adminUser, task.ID, []int64{studentUser.ID})
		require.NoError(t, err)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		tr.EXPECT().IsAssignedToUser(tx, task.ID, studentUser.ID).Return(true, nil).Times(1)
		tr.EXPECT().UnassignFromUser(tx, task.ID, studentUser.ID).Return(nil).Times(1)
		tr.EXPECT().Get(tx, task.ID).Return(task, nil).Times(1)

		err := ts.UnassignFromUsers(tx, teacherUser, task.ID, []int64{studentUser.ID})
		require.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		err := ts.UnassignFromUsers(tx, studentUser, task.ID, []int64{studentUser.ID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestUnAssignTaskFromGroups(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	adminUser := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	teacherUser := schemas.User{ID: 2, Role: types.UserRoleTeacher}
	studentUser := schemas.User{ID: 3, Role: types.UserRoleStudent}
	task := &models.Task{
		ID:        int64(1),
		CreatedBy: teacherUser.ID,
	}

	t.Run("Success with admin", func(t *testing.T) {
		groupID := int64(1)
		tr.EXPECT().Get(tx, task.ID).Return(task, nil).Times(1)
		tr.EXPECT().IsAssignedToGroup(tx, task.ID, groupID).Return(true, nil).Times(1)
		tr.EXPECT().UnassignFromGroup(tx, task.ID, groupID).Return(nil).Times(1)

		err := ts.UnassignFromGroups(tx, adminUser, task.ID, []int64{groupID})
		require.NoError(t, err)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		groupID := int64(1)
		tr.EXPECT().Get(tx, task.ID).Return(task, nil).Times(1)
		tr.EXPECT().IsAssignedToGroup(tx, task.ID, groupID).Return(true, nil).Times(1)
		tr.EXPECT().UnassignFromGroup(tx, task.ID, groupID).Return(nil).Times(1)

		err := ts.UnassignFromGroups(tx, teacherUser, task.ID, []int64{groupID})
		require.NoError(t, err)
	})

	t.Run("Not authorized", func(t *testing.T) {
		groupID := int64(1)
		err := ts.UnassignFromGroups(tx, studentUser, task.ID, []int64{groupID})
		require.ErrorIs(t, err, errors.ErrNotAuthorized)
	})
}

func TestCreateTestCase(t *testing.T) {
	tx := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	teacherUser := schemas.User{ID: 2}
	task := &models.Task{
		ID:        int64(1),
		CreatedBy: teacherUser.ID,
	}

	t.Run("Success", func(t *testing.T) {
		io.EXPECT().DeleteAll(tx, task.ID).Return(nil).Times(1)
		tr.EXPECT().Get(tx, task.ID).Return(task, nil).Times(1)
		io.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).Times(4)
		pathToArchive := createTestArchive(t, "valid")
		err := ts.CreateTestCase(tx, task.ID, pathToArchive)
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		pathToArchive := createTestArchive(t, "valid")
		defer os.Remove(pathToArchive)
		tr.EXPECT().Get(tx, task.ID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := ts.CreateTestCase(tx, task.ID, pathToArchive)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Invalid archive path", func(t *testing.T) {
		tr.EXPECT().Get(tx, task.ID).Return(task, nil).Times(1)

		err := ts.CreateTestCase(tx, task.ID, "INVALIDPATH")
		require.Error(t, err)
	})
}

func TestParseTestCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
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
			pathToArchive := createTestArchive(t, tt.caseType)

			numFiles, err := ts.ParseTestCase(pathToArchive)
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

func TestGetLimits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	tx := &testutils.MockDatabase{}
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	taskID := int64(1)

	teacherUser := schemas.User{ID: 2}
	t.Run("Success", func(t *testing.T) {
		io.EXPECT().GetByTask(tx, taskID).Return([]models.TestCase{{
			ID:          1,
			TaskID:      taskID,
			Order:       1,
			TimeLimit:   10,
			MemoryLimit: 10,
		}}, nil).Times(1)
		tr.EXPECT().Get(tx, taskID).Return(&models.Task{
			ID: 1,
		}, nil).Times(1)

		result, err := ts.GetLimits(tx, teacherUser, taskID)
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestPutLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	tx := &testutils.MockDatabase{}
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr)
	taskID := int64(1)
	ioID := int64(1)
	testCase := &models.TestCase{
		ID:          ioID,
		TaskID:      taskID,
		Order:       1,
		TimeLimit:   10,
		MemoryLimit: 10,
	}

	t.Run("Success", func(t *testing.T) {
		teacherUser := schemas.User{ID: 2}
		io.EXPECT().GetTestCaseID(tx, taskID, testCase.Order).Return(testCase.ID, nil).Times(1)
		io.EXPECT().Get(tx, ioID).Return(testCase, nil).Times(1)
		tr.EXPECT().Get(tx, taskID).Return(&models.Task{
			ID:        taskID,
			CreatedBy: teacherUser.ID,
		}, nil).Times(1)

		newLimits := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{
					Order:       1,
					TimeLimit:   20,
					MemoryLimit: 20,
				},
			},
		}
		expectedModel := &models.TestCase{
			ID:          ioID,
			TaskID:      taskID,
			Order:       testCase.Order,
			TimeLimit:   newLimits.Limits[0].TimeLimit,
			MemoryLimit: newLimits.Limits[0].MemoryLimit,
		}
		io.EXPECT().Put(tx, expectedModel).Return(nil).Times(1)

		err := ts.PutLimits(tx, teacherUser, taskID, newLimits)
		require.NoError(t, err)
	})
}
