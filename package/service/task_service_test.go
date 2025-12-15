package service_test

import (
	"archive/zip"
	stderrors "errors"
	"fmt"
	"os"
	"testing"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/filestorage"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

var (
	studentUser = &schemas.User{ID: 2, Role: types.UserRoleStudent}
	teacherUser = &schemas.User{ID: 3, Role: types.UserRoleTeacher}
	adminUser   = &schemas.User{ID: 1, Role: types.UserRoleAdmin}
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
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, acs)
	t.Run("Success", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		ur.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&models.User{ID: 1, Role: types.UserRoleAdmin}, nil).Times(1)
		tr.EXPECT().GetByTitle(db, task.Title).Return(nil, gorm.ErrRecordNotFound).Times(1)
		tr.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(1), nil).Times(1)
		acs.EXPECT().GrantOwnerAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		taskID, err := ts.Create(db, adminUser, task)
		require.NoError(t, err)
		assert.NotEqual(t, 0, taskID)
	})

	t.Run("Non unique title", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetByTitle(db, task.Title).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		taskID, err := ts.Create(db, adminUser, task)
		require.ErrorIs(t, err, errors.ErrTaskExists)
		assert.Equal(t, int64(0), taskID)
	})

	t.Run("Not authorized", func(t *testing.T) {
		taskID, err := ts.Create(db, studentUser, &schemas.Task{
			Title:     "Test Student Task",
			CreatedBy: studentUser.ID,
		})
		require.ErrorIs(t, err, errors.ErrForbidden)
		assert.Equal(t, int64(0), taskID)
	})
}

func TestGetTaskByTitle(t *testing.T) {
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, nil)

	t.Run("Success", func(t *testing.T) {
		taskID := int64(1)
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetByTitle(db, task.Title).Return(&models.Task{
			ID:        taskID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		taskResp, err := ts.GetByTitle(db, task.Title)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		taskTitle := "Nonexistent Task"
		tr.EXPECT().GetByTitle(db, taskTitle).Return(nil, errors.ErrTaskNotFound).Times(1)
		task, err := ts.GetByTitle(db, taskTitle)
		require.ErrorIs(t, err, errors.ErrTaskNotFound)
		assert.Nil(t, task)
	})
}

func TestGetAllTasks(t *testing.T) {
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, nil)

	paginationParams := schemas.PaginationParams{Limit: 10, Offset: 0, Sort: "id:asc"}

	t.Run("No tasks", func(t *testing.T) {
		tr.EXPECT().GetAll(db,
			paginationParams.Limit,
			paginationParams.Offset,
			paginationParams.Sort,
		).Return([]models.Task{}, int64(0), nil).Times(1)
		tasks, err := ts.GetAll(db, adminUser, paginationParams)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("Success", func(t *testing.T) {
		tasks := []models.Task{
			{
				ID:        1,
				Title:     "Test Task",
				CreatedBy: teacherUser.ID,
				IsVisible: true,
			},
		}
		tr.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tasks, int64(len(tasks)), nil).Times(1)

		resultTasks, err := ts.GetAll(db, studentUser, paginationParams)
		require.NoError(t, err)
		assert.NotEmpty(t, resultTasks)
		assert.True(t, resultTasks[0].IsVisible)
	})

	t.Run("Only globally visible tasks", func(t *testing.T) {
		// Repository should only return globally visible tasks
		tasks := []models.Task{
			{
				ID:        1,
				Title:     "Visible Task",
				CreatedBy: teacherUser.ID,
				IsVisible: true,
			},
		}
		tr.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tasks, int64(len(tasks)), nil).Times(1)

		resultTasks, err := ts.GetAll(db, studentUser, paginationParams)
		require.NoError(t, err)
		assert.Len(t, resultTasks, 1)
		assert.Equal(t, "Visible Task", resultTasks[0].Title)
		assert.True(t, resultTasks[0].IsVisible)
	})
}

func TestGetTask(t *testing.T) {
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	config := testutils.NewTestConfig()
	fs, err := filestorage.NewFileStorageService(config.FileStorageURL)
	require.NoError(t, err)
	ts := service.NewTaskService(fs, fr, tr, io, ur, gr, nil, nil, nil)

	task := &schemas.Task{
		Title:     "Test Task",
		CreatedBy: adminUser.ID,
	}

	t.Run("Success", func(t *testing.T) {
		tr.EXPECT().Get(db, task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		taskResp, err := ts.Get(db, adminUser, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Sucess with assigned task to user", func(t *testing.T) {
		tr.EXPECT().Get(gomock.Any(), task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)

		// tr.EXPECT().IsAssignedToUser(db, task.ID, studentUser.ID).Return(true, nil).Times(1)
		taskResp, err := ts.Get(db, studentUser, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, taskResp.ID)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Sucess with assigned task to group", func(t *testing.T) {
		tr.EXPECT().Get(db, task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)

		taskResp, err := ts.Get(db, studentUser, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, taskResp.ID)
		assert.Equal(t, task.Title, taskResp.Title)
	})

	t.Run("Success with created task", func(t *testing.T) {
		tr.EXPECT().Get(db, task.ID).Return(&models.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(1)
		taskResp, err := ts.Get(db, teacherUser, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, taskResp.ID)
		assert.Equal(t, task.Title, taskResp.Title)
		assert.Equal(t, task.CreatedBy, taskResp.CreatedBy)
	})

	t.Run("Fail with non existent task", func(t *testing.T) {
		taskID := int64(0)
		tr.EXPECT().Get(db, taskID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		taskResp, err := ts.Get(db, adminUser, taskID)
		require.ErrorIs(t, err, errors.ErrNotFound)
		assert.Nil(t, taskResp)
	})
}

func TestDeleteTask(t *testing.T) {
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, acs)
	taskID := int64(1)

	t.Run("Success for admin", func(t *testing.T) {
		tr.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		tr.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&models.Task{
			ID: taskID,
		}, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		err := ts.Delete(db, adminUser, taskID)
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		tr.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(gorm.ErrRecordNotFound).Times(1)
		tr.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&models.Task{
			ID: taskID,
		}, nil).Times(1)
		err := ts.Delete(db, adminUser, 0)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Not authorized", func(t *testing.T) {
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrForbidden).Times(1)
		tr.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&models.Task{
			ID: taskID,
		}, nil).Times(1)
		err := ts.Delete(db, studentUser, taskID)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})
}

func TestEditTask(t *testing.T) {
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, acs)
	taskID := int64(1)

	t.Run("Success", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().Get(db, taskID).Return(&models.Task{
			ID:        taskID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
		}, nil).Times(2)
		tr.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		newTitle := "Updated Task"
		updatedTask := &schemas.EditTask{Title: &newTitle}
		err := ts.Edit(db, adminUser, taskID, updatedTask)
		require.NoError(t, err)
	})

	t.Run("Nonexistent task", func(t *testing.T) {
		newTitle := "Updated Task"
		updatedTask := &schemas.EditTask{Title: &newTitle}
		tr.EXPECT().Get(db, int64(0)).Return(nil, errors.ErrTaskNotFound).Times(1)
		err := ts.Edit(db, adminUser, 0, updatedTask)
		require.ErrorIs(t, err, errors.ErrTaskNotFound)
	})

	t.Run("Update isVisible", func(t *testing.T) {
		task := &models.Task{
			ID:        taskID,
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
			IsVisible: true,
		}
		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(2)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		tr.EXPECT().Edit(db, taskID, gomock.Any()).DoAndReturn(func(db database.Database, id int64, updatedTask *models.Task) error {
			// Verify that IsVisible was updated
			assert.False(t, updatedTask.IsVisible)
			return nil
		}).Times(1)

		isVisible := false
		updatedTask := &schemas.EditTask{IsVisible: &isVisible}
		err := ts.Edit(db, adminUser, taskID, updatedTask)
		require.NoError(t, err)
	})
}

func TestGetAllCreatedTasks(t *testing.T) {
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, nil)
	taskID := int64(1)
	queryParams := schemas.PaginationParams{Limit: 10, Offset: 0, Sort: "id:asc"}

	t.Run("No tasks", func(t *testing.T) {
		tr.EXPECT().GetAllCreated(
			db,
			adminUser.ID,
			queryParams.Offset,
			queryParams.Limit,
			queryParams.Sort,
		).Return([]models.Task{}, int64(0), nil).Times(1)
		result, err := ts.GetAllCreated(db, adminUser, queryParams)
		require.NoError(t, err)
		assert.Empty(t, result.Items)
		assert.Equal(t, 0, result.Pagination.TotalItems)
	})

	t.Run("Success with admin", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetAllCreated(
			db,
			adminUser.ID,
			queryParams.Offset,
			queryParams.Limit,
			queryParams.Sort,
		).Return([]models.Task{
			{
				ID:        taskID,
				Title:     task.Title,
				CreatedBy: task.CreatedBy,
			},
		}, int64(1), nil).Times(1)
		result, err := ts.GetAllCreated(db, adminUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Items)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, task.Title, result.Items[0].Title)
		assert.Equal(t, task.CreatedBy, result.Items[0].CreatedBy)
		assert.Equal(t, 1, result.Pagination.TotalItems)
	})

	t.Run("Success with teacher", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Teacher Task",
			CreatedBy: teacherUser.ID,
		}
		tr.EXPECT().GetAllCreated(
			db,
			teacherUser.ID,
			queryParams.Offset,
			queryParams.Limit,
			queryParams.Sort,
		).Return([]models.Task{
			{
				ID:        taskID,
				Title:     task.Title,
				CreatedBy: task.CreatedBy,
			},
		}, int64(1), nil).Times(1)
		result, err := ts.GetAllCreated(db, teacherUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Items)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, task.Title, result.Items[0].Title)
		assert.Equal(t, task.CreatedBy, result.Items[0].CreatedBy)
		assert.Equal(t, 1, result.Pagination.TotalItems)
	})

	t.Run("Different teachers", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Teacher Task 2",
			CreatedBy: teacherUser.ID,
		}
		tr.EXPECT().GetAllCreated(
			db,
			teacherUser.ID,
			queryParams.Offset,
			queryParams.Limit,
			queryParams.Sort,
		).Return([]models.Task{
			{
				ID:        taskID,
				Title:     task.Title,
				CreatedBy: task.CreatedBy,
			},
		}, int64(1), nil).Times(1)
		result, err := ts.GetAllCreated(db, teacherUser, queryParams)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Items)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, task.Title, result.Items[0].Title)
		assert.Equal(t, task.CreatedBy, result.Items[0].CreatedBy)
		assert.Equal(t, 1, result.Pagination.TotalItems)
	})

	t.Run("Not authorized", func(t *testing.T) {
		result, err := ts.GetAllCreated(db, studentUser, queryParams)
		require.ErrorIs(t, err, errors.ErrForbidden)
		assert.Empty(t, result.Items)
		assert.Equal(t, 0, result.Pagination.TotalItems)
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
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, nil)
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
	db := &testutils.MockDatabase{}
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, nil)
	taskID := int64(1)

	t.Run("Success", func(t *testing.T) {
		io.EXPECT().GetByTask(db, taskID).Return([]models.TestCase{{
			ID:            1,
			TaskID:        taskID,
			Order:         1,
			TimeLimitMs:   10,
			MemoryLimitKB: 10,
		}}, nil).Times(1)
		tr.EXPECT().Get(db, taskID).Return(&models.Task{
			ID: 1,
		}, nil).Times(1)

		result, err := ts.GetLimits(db, teacherUser, taskID)
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
	acs := mock_service.NewMockAccessControlService(ctrl)
	db := &testutils.MockDatabase{}
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, acs)
	taskID := int64(1)
	ioID := int64(1)
	testCase := &models.TestCase{
		ID:            ioID,
		TaskID:        taskID,
		Order:         1,
		TimeLimitMs:   10,
		MemoryLimitKB: 10,
	}

	t.Run("Success", func(t *testing.T) {
		io.EXPECT().GetTestCaseID(db, taskID, testCase.Order).Return(testCase.ID, nil).Times(1)
		io.EXPECT().Get(db, ioID).Return(testCase, nil).Times(1)
		tr.EXPECT().Get(db, taskID).Return(&models.Task{
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
			ID:            ioID,
			TaskID:        taskID,
			Order:         testCase.Order,
			TimeLimitMs:   newLimits.Limits[0].TimeLimit,
			MemoryLimitKB: newLimits.Limits[0].MemoryLimit,
		}
		io.EXPECT().Put(db, expectedModel).Return(nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		err := ts.PutLimits(db, teacherUser, taskID, newLimits)
		require.NoError(t, err)
	})
}

func TestGetMyLiveTasks(t *testing.T) {
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tr := mock_repository.NewMockTaskRepository(ctrl)
	cr := mock_repository.NewMockContestRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, sr, cr, nil)

	paginationParams := schemas.PaginationParams{Limit: 10, Offset: 0}

	t.Run("Success with no tasks", func(t *testing.T) {
		tr.EXPECT().GetLiveAssignedTasksGroupedByContest(db, studentUser.ID, paginationParams.Limit, paginationParams.Offset).Return(map[int64][]models.Task{}, nil).Times(1)

		result, err := ts.GetMyLiveTasks(db, studentUser, paginationParams)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Contests)
	})

	t.Run("Success with tasks in one contest", func(t *testing.T) {
		contestID := int64(1)
		taskID := int64(1)
		contestTasksMap := map[int64][]models.Task{
			contestID: {
				{ID: taskID, Title: "Task 1", CreatedBy: teacherUser.ID},
			},
		}
		contest := &models.Contest{
			ID:   contestID,
			Name: "Test Contest",
		}

		tr.EXPECT().GetLiveAssignedTasksGroupedByContest(db, studentUser.ID, paginationParams.Limit, paginationParams.Offset).Return(contestTasksMap, nil).Times(1)
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		sr.EXPECT().GetAttemptCountForTaskByUser(db, taskID, studentUser.ID).Return(2, nil).Times(1)
		sr.EXPECT().GetBestScoreForTaskByUser(db, taskID, studentUser.ID).Return(float64(80.0), nil).Times(1)

		result, err := ts.GetMyLiveTasks(db, studentUser, paginationParams)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Contests, 1)
		assert.Equal(t, contestID, result.Contests[0].ContestID)
		assert.Equal(t, "Test Contest", result.Contests[0].ContestName)
		assert.Len(t, result.Contests[0].Tasks, 1)
		assert.Equal(t, taskID, result.Contests[0].Tasks[0].ID)
		assert.Equal(t, 2, result.Contests[0].Tasks[0].AttemptsSummary.AttemptCount)
		assert.InEpsilon(t, float64(80.0), result.Contests[0].Tasks[0].AttemptsSummary.BestScore, 0.001)
	})

	t.Run("Success with no attempts", func(t *testing.T) {
		contestID := int64(1)
		taskID := int64(1)
		contestTasksMap := map[int64][]models.Task{
			contestID: {
				{ID: taskID, Title: "Task 1", CreatedBy: teacherUser.ID},
			},
		}
		contest := &models.Contest{
			ID:   contestID,
			Name: "Test Contest",
		}

		tr.EXPECT().GetLiveAssignedTasksGroupedByContest(db, studentUser.ID, paginationParams.Limit, paginationParams.Offset).Return(contestTasksMap, nil).Times(1)
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		sr.EXPECT().GetAttemptCountForTaskByUser(db, taskID, studentUser.ID).Return(0, nil).Times(1)
		// GetBestScoreForTaskByUser should NOT be called when attemptCount is 0

		result, err := ts.GetMyLiveTasks(db, studentUser, paginationParams)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Contests, 1)
		assert.Equal(t, 0, result.Contests[0].Tasks[0].AttemptsSummary.AttemptCount)
		assert.InDelta(t, float64(0), result.Contests[0].Tasks[0].AttemptsSummary.BestScore, 0.001)
	})

	t.Run("Error getting live tasks", func(t *testing.T) {
		tr.EXPECT().GetLiveAssignedTasksGroupedByContest(db, studentUser.ID, paginationParams.Limit, paginationParams.Offset).Return(nil, gorm.ErrInvalidDB).Times(1)

		result, err := ts.GetMyLiveTasks(db, studentUser, paginationParams)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error getting contest skips contest", func(t *testing.T) {
		contestID := int64(1)
		taskID := int64(1)
		contestTasksMap := map[int64][]models.Task{
			contestID: {
				{ID: taskID, Title: "Task 1", CreatedBy: teacherUser.ID},
			},
		}

		tr.EXPECT().GetLiveAssignedTasksGroupedByContest(db, studentUser.ID, paginationParams.Limit, paginationParams.Offset).Return(contestTasksMap, nil).Times(1)
		cr.EXPECT().Get(db, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		result, err := ts.GetMyLiveTasks(db, studentUser, paginationParams)
		require.NoError(t, err)
		assert.NotNil(t, result)
		// Contest should be skipped due to error
		assert.Empty(t, result.Contests)
	})

	t.Run("Error getting attempt count skips task", func(t *testing.T) {
		contestID := int64(1)
		taskID := int64(1)
		contestTasksMap := map[int64][]models.Task{
			contestID: {
				{ID: taskID, Title: "Task 1", CreatedBy: teacherUser.ID},
			},
		}
		contest := &models.Contest{
			ID:   contestID,
			Name: "Test Contest",
		}

		tr.EXPECT().GetLiveAssignedTasksGroupedByContest(db, studentUser.ID, paginationParams.Limit, paginationParams.Offset).Return(contestTasksMap, nil).Times(1)
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		sr.EXPECT().GetAttemptCountForTaskByUser(db, taskID, studentUser.ID).Return(0, gorm.ErrInvalidDB).Times(1)

		result, err := ts.GetMyLiveTasks(db, studentUser, paginationParams)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Contests, 1)
		// Task should be skipped due to error
		assert.Empty(t, result.Contests[0].Tasks)
	})

	t.Run("Error getting best score skips task", func(t *testing.T) {
		contestID := int64(1)
		taskID := int64(1)
		contestTasksMap := map[int64][]models.Task{
			contestID: {
				{ID: taskID, Title: "Task 1", CreatedBy: teacherUser.ID},
			},
		}
		contest := &models.Contest{
			ID:   contestID,
			Name: "Test Contest",
		}

		tr.EXPECT().GetLiveAssignedTasksGroupedByContest(db, studentUser.ID, paginationParams.Limit, paginationParams.Offset).Return(contestTasksMap, nil).Times(1)
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		sr.EXPECT().GetAttemptCountForTaskByUser(db, taskID, studentUser.ID).Return(2, nil).Times(1)
		sr.EXPECT().GetBestScoreForTaskByUser(db, taskID, studentUser.ID).Return(float64(0), gorm.ErrInvalidDB).Times(1)

		result, err := ts.GetMyLiveTasks(db, studentUser, paginationParams)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Contests, 1)
		// Task should be skipped due to error
		assert.Empty(t, result.Contests[0].Tasks)
	})
}

func TestProcessAndUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := &testutils.MockDatabase{}
	fsMock := filestorage.NewMockFileStorageService(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)

	ts := service.NewTaskService(fsMock, fr, tr, io, ur, gr, nil, nil, acs)

	taskID := int64(1)
	archivePath := "/test/archive.zip"
	task := &models.Task{
		ID:        taskID,
		Title:     "Test Task",
		CreatedBy: teacherUser.ID,
	}

	t.Run("Success", func(t *testing.T) {
		uploadedFiles := &filestorage.UploadedTaskFiles{
			DescriptionFile: filestorage.UploadedFile{
				Filename:   "description.pdf",
				Path:       "task/1/description.pdf",
				Bucket:     "maxit",
				ServerType: "filestorage",
			},
			InputFiles: []filestorage.UploadedFile{
				{Filename: "1.in", Path: "task/1/input/1.in", Bucket: "maxit", ServerType: "filestorage"},
			},
			OutputFiles: []filestorage.UploadedFile{
				{Filename: "1.out", Path: "task/1/output/1.out", Bucket: "maxit", ServerType: "filestorage"},
			},
		}

		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(2)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(uploadedFiles, nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(nil).Times(3) // description, input, output files
		tr.EXPECT().Edit(db, taskID, gomock.Any()).Return(nil).Times(1)
		io.EXPECT().Create(db, gomock.Any()).Return(nil).Times(1)

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.NoError(t, err)
	})

	t.Run("Permission denied", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrForbidden).Times(1)

		err := ts.ProcessAndUpload(db, studentUser, taskID, archivePath)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("Task not found", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Invalid archive structure", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(stderrors.New("invalid archive")).Times(1)

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate archive structure")
	})

	t.Run("Upload task error", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(nil, stderrors.New("upload failed")).Times(1)

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to upload task archive")
	})

	t.Run("Save description file error", func(t *testing.T) {
		uploadedFiles := &filestorage.UploadedTaskFiles{
			DescriptionFile: filestorage.UploadedFile{
				Filename: "description.pdf",
				Path:     "task/1/description.pdf",
			},
			InputFiles:  []filestorage.UploadedFile{},
			OutputFiles: []filestorage.UploadedFile{},
		}

		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(uploadedFiles, nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save description file")
	})

	t.Run("Get task error after upload", func(t *testing.T) {
		uploadedFiles := &filestorage.UploadedTaskFiles{
			DescriptionFile: filestorage.UploadedFile{
				Filename: "description.pdf",
				Path:     "task/1/description.pdf",
			},
			InputFiles:  []filestorage.UploadedFile{},
			OutputFiles: []filestorage.UploadedFile{},
		}

		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(uploadedFiles, nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(nil).Times(1)
		tr.EXPECT().Get(db, taskID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Update task error", func(t *testing.T) {
		uploadedFiles := &filestorage.UploadedTaskFiles{
			DescriptionFile: filestorage.UploadedFile{
				Filename: "description.pdf",
				Path:     "task/1/description.pdf",
			},
			InputFiles:  []filestorage.UploadedFile{},
			OutputFiles: []filestorage.UploadedFile{},
		}

		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(2)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(uploadedFiles, nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(nil).Times(1)
		tr.EXPECT().Edit(db, taskID, gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update task with description file")
	})

	t.Run("Save input file error", func(t *testing.T) {
		uploadedFiles := &filestorage.UploadedTaskFiles{
			DescriptionFile: filestorage.UploadedFile{
				Filename: "description.pdf",
				Path:     "task/1/description.pdf",
			},
			InputFiles: []filestorage.UploadedFile{
				{Filename: "1.in", Path: "task/1/input/1.in"},
			},
			OutputFiles: []filestorage.UploadedFile{
				{Filename: "1.out", Path: "task/1/output/1.out"},
			},
		}

		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(2)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(uploadedFiles, nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(nil).Times(1) // description file
		tr.EXPECT().Edit(db, taskID, gomock.Any()).Return(nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(gorm.ErrInvalidDB).Times(1) // input file

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save input file")
	})

	t.Run("Save output file error", func(t *testing.T) {
		uploadedFiles := &filestorage.UploadedTaskFiles{
			DescriptionFile: filestorage.UploadedFile{
				Filename: "description.pdf",
				Path:     "task/1/description.pdf",
			},
			InputFiles: []filestorage.UploadedFile{
				{Filename: "1.in", Path: "task/1/input/1.in"},
			},
			OutputFiles: []filestorage.UploadedFile{
				{Filename: "1.out", Path: "task/1/output/1.out"},
			},
		}

		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(2)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(uploadedFiles, nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(nil).Times(2) // description file + input file
		tr.EXPECT().Edit(db, taskID, gomock.Any()).Return(nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(gorm.ErrInvalidDB).Times(1) // output file

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save output file")
	})

	t.Run("Create test case error", func(t *testing.T) {
		uploadedFiles := &filestorage.UploadedTaskFiles{
			DescriptionFile: filestorage.UploadedFile{
				Filename: "description.pdf",
				Path:     "task/1/description.pdf",
			},
			InputFiles: []filestorage.UploadedFile{
				{Filename: "1.in", Path: "task/1/input/1.in"},
			},
			OutputFiles: []filestorage.UploadedFile{
				{Filename: "1.out", Path: "task/1/output/1.out"},
			},
		}

		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(2)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(uploadedFiles, nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(nil).Times(3) // description + input + output files
		tr.EXPECT().Edit(db, taskID, gomock.Any()).Return(nil).Times(1)
		io.EXPECT().Create(db, gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save input output")
	})

	t.Run("Success with multiple input output files", func(t *testing.T) {
		uploadedFiles := &filestorage.UploadedTaskFiles{
			DescriptionFile: filestorage.UploadedFile{
				Filename: "description.pdf",
				Path:     "task/1/description.pdf",
			},
			InputFiles: []filestorage.UploadedFile{
				{Filename: "1.in", Path: "task/1/input/1.in"},
				{Filename: "2.in", Path: "task/1/input/2.in"},
				{Filename: "3.in", Path: "task/1/input/3.in"},
			},
			OutputFiles: []filestorage.UploadedFile{
				{Filename: "1.out", Path: "task/1/output/1.out"},
				{Filename: "2.out", Path: "task/1/output/2.out"},
				{Filename: "3.out", Path: "task/1/output/3.out"},
			},
		}

		tr.EXPECT().Get(db, taskID).Return(task, nil).Times(2)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		fsMock.EXPECT().ValidateArchiveStructure(archivePath).Return(nil).Times(1)
		fsMock.EXPECT().UploadTask(taskID, archivePath).Return(uploadedFiles, nil).Times(1)
		fr.EXPECT().Create(db, gomock.Any()).Return(nil).Times(7) // 1 description + 3 input + 3 output
		tr.EXPECT().Edit(db, taskID, gomock.Any()).Return(nil).Times(1)
		io.EXPECT().Create(db, gomock.Any()).Return(nil).Times(3) // 3 test cases

		err := ts.ProcessAndUpload(db, teacherUser, taskID, archivePath)
		require.NoError(t, err)
	})
}

func TestGetLimitsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, nil)
	taskID := int64(1)

	t.Run("Task not found", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		result, err := ts.GetLimits(db, teacherUser, taskID)
		require.ErrorIs(t, err, errors.ErrNotFound)
		assert.Nil(t, result)
	})

	t.Run("Error getting test cases", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(&models.Task{ID: taskID}, nil).Times(1)
		io.EXPECT().GetByTask(db, taskID).Return(nil, gorm.ErrInvalidDB).Times(1)

		result, err := ts.GetLimits(db, teacherUser, taskID)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestPutLimitsErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, acs)
	taskID := int64(1)

	t.Run("Permission denied", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(&models.Task{ID: taskID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrForbidden).Times(1)

		newLimits := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 20, MemoryLimit: 20},
			},
		}
		err := ts.PutLimits(db, studentUser, taskID, newLimits)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("Task not found", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		newLimits := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 20, MemoryLimit: 20},
			},
		}
		err := ts.PutLimits(db, teacherUser, taskID, newLimits)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Error getting test case ID", func(t *testing.T) {
		tr.EXPECT().Get(db, taskID).Return(&models.Task{ID: taskID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		io.EXPECT().GetTestCaseID(db, taskID, 1).Return(int64(0), gorm.ErrRecordNotFound).Times(1)

		newLimits := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 20, MemoryLimit: 20},
			},
		}
		err := ts.PutLimits(db, teacherUser, taskID, newLimits)
		require.Error(t, err)
	})

	t.Run("Error getting test case", func(t *testing.T) {
		ioID := int64(1)
		tr.EXPECT().Get(db, taskID).Return(&models.Task{ID: taskID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		io.EXPECT().GetTestCaseID(db, taskID, 1).Return(ioID, nil).Times(1)
		io.EXPECT().Get(db, ioID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		newLimits := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 20, MemoryLimit: 20},
			},
		}
		err := ts.PutLimits(db, teacherUser, taskID, newLimits)
		require.Error(t, err)
	})

	t.Run("Error putting test case", func(t *testing.T) {
		ioID := int64(1)
		testCase := &models.TestCase{
			ID:            ioID,
			TaskID:        taskID,
			Order:         1,
			TimeLimitMs:   10,
			MemoryLimitKB: 10,
		}
		tr.EXPECT().Get(db, taskID).Return(&models.Task{ID: taskID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		io.EXPECT().GetTestCaseID(db, taskID, 1).Return(ioID, nil).Times(1)
		io.EXPECT().Get(db, ioID).Return(testCase, nil).Times(1)
		io.EXPECT().Put(db, gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)

		newLimits := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 20, MemoryLimit: 20},
			},
		}
		err := ts.PutLimits(db, teacherUser, taskID, newLimits)
		require.Error(t, err)
	})
}

func TestCreateTaskErrors(t *testing.T) {
	db := &testutils.MockDatabase{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ur := mock_repository.NewMockUserRepository(ctrl)
	fr := mock_repository.NewMockFile(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	io := mock_repository.NewMockTestCaseRepository(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	ts := service.NewTaskService(nil, fr, tr, io, ur, gr, nil, nil, acs)

	t.Run("Error getting task by title", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetByTitle(db, task.Title).Return(nil, gorm.ErrInvalidDB).Times(1)

		taskID, err := ts.Create(db, adminUser, task)
		require.Error(t, err)
		assert.Equal(t, int64(0), taskID)
	})

	t.Run("Error getting user", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetByTitle(db, task.Title).Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Get(gomock.Any(), adminUser.ID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		taskID, err := ts.Create(db, adminUser, task)
		require.Error(t, err)
		assert.Equal(t, int64(0), taskID)
	})

	t.Run("Error creating task", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetByTitle(db, task.Title).Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Get(gomock.Any(), adminUser.ID).Return(&models.User{ID: adminUser.ID}, nil).Times(1)
		tr.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(0), gorm.ErrInvalidDB).Times(1)

		taskID, err := ts.Create(db, adminUser, task)
		require.Error(t, err)
		assert.Equal(t, int64(0), taskID)
	})

	t.Run("Error granting owner access", func(t *testing.T) {
		task := &schemas.Task{
			Title:     "Test Task",
			CreatedBy: adminUser.ID,
		}
		tr.EXPECT().GetByTitle(db, task.Title).Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Get(gomock.Any(), adminUser.ID).Return(&models.User{ID: adminUser.ID}, nil).Times(1)
		tr.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(1), nil).Times(1)
		acs.EXPECT().GrantOwnerAccess(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)

		taskID, err := ts.Create(db, adminUser, task)
		require.Error(t, err)
		assert.Equal(t, int64(-1), taskID)
	})
}
