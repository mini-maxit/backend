package service_test

import (
	"encoding/json"
	"github.com/mini-maxit/backend/internal/database"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/filestorage"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

// testSetup holds all mocks and the service for testing
type testSetup struct {
	ctrl                       *gomock.Controller
	contestService             *mock_service.MockContestService
	fileRepository             *mock_repository.MockFile
	submissionRepository       *mock_repository.MockSubmissionRepository
	submissionResultRepository *mock_repository.MockSubmissionResultRepository
	testCaseRepository         *mock_repository.MockTestCaseRepository
	testResultRepository       *mock_repository.MockTestRepository
	groupRepository            *mock_repository.MockGroupRepository
	taskRepository             *mock_repository.MockTaskRepository
	languageService            *mock_service.MockLanguageService
	taskService                *mock_service.MockTaskService
	userService                *mock_service.MockUserService
	queueService               *mock_service.MockQueueService
	service                    service.SubmissionService
}

// setupSubmissionServiceTest initializes all mocks and returns a testSetup struct
func setupSubmissionServiceTest(t *testing.T) *testSetup {
	ctrl := gomock.NewController(t)

	contestService := mock_service.NewMockContestService(ctrl)
	fileRepository := mock_repository.NewMockFile(ctrl)
	submissionRepository := mock_repository.NewMockSubmissionRepository(ctrl)
	submissionResultRepository := mock_repository.NewMockSubmissionResultRepository(ctrl)
	inputOutputRepository := mock_repository.NewMockTestCaseRepository(ctrl)
	testResultRepository := mock_repository.NewMockTestRepository(ctrl)
	groupRepository := mock_repository.NewMockGroupRepository(ctrl)
	taskRepository := mock_repository.NewMockTaskRepository(ctrl)
	languageService := mock_service.NewMockLanguageService(ctrl)
	taskService := mock_service.NewMockTaskService(ctrl)
	userService := mock_service.NewMockUserService(ctrl)
	queueService := mock_service.NewMockQueueService(ctrl)
	fs, err := filestorage.NewFileStorageService("dummy")
	require.NoError(t, err)

	svc := service.NewSubmissionService(
		contestService,
		fs,
		fileRepository,
		submissionRepository,
		submissionResultRepository,
		inputOutputRepository,
		testResultRepository,
		groupRepository,
		taskRepository,
		languageService,
		taskService,
		userService,
		queueService,
	)

	return &testSetup{
		ctrl:                       ctrl,
		contestService:             contestService,
		fileRepository:             fileRepository,
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
		testCaseRepository:         inputOutputRepository,
		testResultRepository:       testResultRepository,
		groupRepository:            groupRepository,
		taskRepository:             taskRepository,
		languageService:            languageService,
		taskService:                taskService,
		userService:                userService,
		queueService:               queueService,
		service:                    svc,
	}
}

func TestCreate(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()

	submission := &models.Submission{
		TaskID: 1,
		UserID: 1,

		Order:      1,
		LanguageID: 1,
		Status:     types.SubmissionStatusReceived,
	}

	t.Run("Succesful", func(t *testing.T) {
		submissionID := rand.Int64()
		expectedModel := &models.Submission{
			TaskID:     rand.Int64(),
			UserID:     rand.Int64(),
			Order:      rand.Int(),
			LanguageID: rand.Int64(),
			ContestID:  nil,
			FileID:     rand.Int64(),
			Status:     types.SubmissionStatusReceived,
		}
		setup.submissionRepository.EXPECT().Create(
			gomock.Any(),
			gomock.AssignableToTypeOf(&models.Submission{}),
		).DoAndReturn(func(_ *database.DB, s *models.Submission) (int64, error) {
			if !reflect.DeepEqual(s, expectedModel) {
				t.Fatalf("invalid submission model passed to repo. exp=%v got=%v", expectedModel, s)
			}
			submission.ID = submissionID
			submission.SubmittedAt = time.Now()
			return submission.ID, nil
		}).Times(1)

		id, err := setup.service.Create(nil, expectedModel.TaskID, expectedModel.UserID, expectedModel.LanguageID, expectedModel.ContestID, expectedModel.Order, expectedModel.FileID)
		require.NoError(t, err)
		assert.Equal(t, submissionID, id)
	})

	t.Run("Fails if repository fails unexpectedly", func(t *testing.T) {
		setup.submissionRepository.EXPECT().
			Create(gomock.Any(), gomock.AssignableToTypeOf(&models.Submission{})).
			DoAndReturn(func(_ *database.DB, _ *models.Submission) (int64, error) {
				return 0, gorm.ErrInvalidData
			}).
			Times(1)
		contestID := int64(1)
		id, err := setup.service.Create(nil, 1, 1, 1, &contestID, 1, 1)
		require.Error(t, err)
		assert.Equal(t, int64(0), id)
	})
}

func TestCreateSubmissionResult(t *testing.T) {
	submission := &models.Submission{
		ID:          1,
		TaskID:      1,
		UserID:      1,
		Order:       1,
		LanguageID:  1,
		Status:      types.SubmissionStatusReceived,
		SubmittedAt: time.Now(),
	}

	testCases := []struct {
		name          string
		setupMocks    func(*mock_repository.MockSubmissionRepository, *mock_repository.MockSubmissionResultRepository, *mock_repository.MockTestCaseRepository, *mock_repository.MockTestRepository)
		expectedID    int64
		expectedErr   bool
		queueResponse schemas.QueueResponseMessage
	}{
		{
			name: "Could not unmarshal response message payload",
			setupMocks: func(mSubmissionRepo *mock_repository.MockSubmissionRepository, mSubmissionResRepo *mock_repository.MockSubmissionResultRepository, mIORepo *mock_repository.MockTestCaseRepository, mTestRepo *mock_repository.MockTestRepository) {
				// No mocks needed for this test case
			},
			expectedID:    int64(-1),
			expectedErr:   true,
			queueResponse: schemas.QueueResponseMessage{},
		},
		{
			name: "Could not put test result",
			setupMocks: func(mSubmissionRepo *mock_repository.MockSubmissionRepository, mSubmissionResRepo *mock_repository.MockSubmissionResultRepository, mIORepo *mock_repository.MockTestCaseRepository, mTestRepo *mock_repository.MockTestRepository) {
				submissionResult := &models.SubmissionResult{ID: int64(1), SubmissionID: int64(1)}
				test1 := &models.TestResult{}
				mSubmissionResRepo.EXPECT().GetBySubmission(gomock.Any(), int64(1)).Return(submissionResult, nil).Times(1)
				mSubmissionResRepo.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mTestRepo.EXPECT().GetBySubmissionAndOrder(gomock.Any(), int64(1), 1).Return(test1, nil).Times(1)
				mTestRepo.EXPECT().Put(gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidData).Times(1)
			},
			expectedID:  int64(-1),
			expectedErr: true,
			queueResponse: schemas.QueueResponseMessage{Payload: json.RawMessage(`
				{
				"status_code":1,
				"message":"solution executed successfully",
				"test_results":
				[
					{"passed":false,"status_code": 1, "error_message":"difference at line 1:\noutput:   hello, world!\nexpected: hello world!\n\n",
					"order":1},
					{"passed":true,"status_code": 0, "error_message":"","order":2}
				]
				}
			`)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupSubmissionServiceTest(t)
			defer setup.ctrl.Finish()

			tc.setupMocks(setup.submissionRepository, setup.submissionResultRepository, setup.testCaseRepository, setup.testResultRepository)
			id, err := setup.service.CreateSubmissionResult(nil, submission.ID, tc.queueResponse)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedID, id)
		})
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()

	t.Run("Success", func(t *testing.T) {
		expectedLanguages := []schemas.LanguageConfig{
			{Type: "Python", Version: "3.8"},
			{Type: "Python", Version: "3.9"},
			{Type: "Go", Version: "1.22"},
			{Type: "Go", Version: "1.23"},
		}

		setup.languageService.EXPECT().GetAllEnabled(gomock.Any()).Return(expectedLanguages, nil).Times(1)

		languages, err := setup.service.GetAvailableLanguages(nil)

		require.NoError(t, err)
		assert.Equal(t, expectedLanguages, languages)
	})

	t.Run("Error fetching languages", func(t *testing.T) {
		setup.languageService.EXPECT().GetAllEnabled(gomock.Any()).Return(nil, gorm.ErrRecordNotFound).Times(1)

		languages, err := setup.service.GetAvailableLanguages(nil)

		require.Error(t, err)
		assert.Nil(t, languages)
	})
}
func TestGetAll(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()
	testCases := []struct {
		name           string
		user           schemas.User
		userID         *int64
		expectedMethod func() *gomock.Call
		expectedResult []schemas.Submission
		expectedErr    bool
	}{
		{
			name:   "Admin retrieves all submissions",
			user:   schemas.User{Role: "admin"},
			userID: nil,
			expectedMethod: func() *gomock.Call {
				return setup.submissionRepository.EXPECT().GetAll(gomock.Any(), 10, 0, "submitted_at:desc").Return([]models.Submission{
					{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
					{ID: 2, TaskID: 2, UserID: 2, Status: types.SubmissionStatusEvaluated},
				}, nil).Times(1)
			},
			expectedResult: []schemas.Submission{
				{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
				{ID: 2, TaskID: 2, UserID: 2, Status: types.SubmissionStatusEvaluated},
			},
			expectedErr: false,
		},
		{
			name:   "Teacher retrieves submissions for their tasks",
			user:   schemas.User{Role: "teacher", ID: 1},
			userID: nil,
			expectedMethod: func() *gomock.Call {
				return setup.submissionRepository.EXPECT().GetAllForTeacher(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
					[]models.Submission{
						{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
						{ID: 2, TaskID: 2, UserID: 2, Status: types.SubmissionStatusEvaluated},
					}, nil).Times(1)
			},
			expectedResult: []schemas.Submission{
				{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
				{ID: 2, TaskID: 2, UserID: 2, Status: types.SubmissionStatusEvaluated},
			},
			expectedErr: false,
		},
		{
			name:   "Student retrieves their own submissions",
			user:   schemas.User{Role: "student", ID: 1},
			userID: nil,
			expectedMethod: func() *gomock.Call {
				return setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
					[]models.Submission{
						{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
					}, nil).Times(1)
			},
			expectedResult: []schemas.Submission{
				{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
			},
			expectedErr: false,
		},
		{
			name:   "Error retrieving submissions",
			user:   schemas.User{Role: "admin"},
			userID: nil,
			expectedMethod: func() *gomock.Call {
				return setup.submissionRepository.EXPECT().GetAll(gomock.Any(), 10, 0, "submitted_at:desc").Return(
					nil, gorm.ErrInvalidData,
				).Times(1)
			},
			expectedResult: nil,
			expectedErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.expectedMethod()
			queryParams := schemas.PaginationParams{Limit: 10, Offset: 0, Sort: ""}

			submissions, err := setup.service.GetAll(nil, tc.user, tc.userID, nil, nil, queryParams)

			if tc.expectedErr {
				require.Error(t, err)
				assert.Nil(t, submissions)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, submissions)
			}
		})
	}
}
func TestGet(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()

	testCases := []struct {
		name               string
		user               schemas.User
		expectedSubmission *models.Submission
		expectedErr        bool
	}{
		{
			name:               "Admin retrieves a submission",
			user:               schemas.User{Role: "admin"},
			expectedSubmission: &models.Submission{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
			expectedErr:        false,
		},
		{
			name:               "Student tries to access another user's submission",
			user:               schemas.User{Role: "student", ID: 1},
			expectedSubmission: &models.Submission{ID: 1, TaskID: 1, UserID: 2, Status: types.SubmissionStatusReceived},
			expectedErr:        true,
		},
		{
			name: "Teacher tries to access a submission for a task they didn't create",
			user: schemas.User{Role: "teacher", ID: 2},
			expectedSubmission: &models.Submission{
				ID:     1,
				TaskID: 1,
				UserID: 2,
				Task:   models.Task{CreatedBy: 3},
			},
			expectedErr: true,
		},
		{
			name: "Teacher retrieves a submission for a task they created",
			user: schemas.User{Role: "teacher", ID: 2},
			expectedSubmission: &models.Submission{
				ID:     1,
				TaskID: 1,
				UserID: 2,
				Task:   models.Task{CreatedBy: 2},
			},
			expectedErr: false,
		},
		{
			name:               "Error retrieving submission",
			user:               schemas.User{Role: "admin"},
			expectedSubmission: nil,
			expectedErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedSubmission != nil {
				setup.submissionRepository.EXPECT().Get(gomock.Any(), int64(1)).Return(tc.expectedSubmission, nil).Times(1)
			} else {
				setup.submissionRepository.EXPECT().Get(gomock.Any(), int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)
			}

			submission, err := setup.service.Get(nil, 1, tc.user)

			if tc.expectedErr {
				require.Error(t, err)
				assert.Equal(t, schemas.Submission{}, submission)
			} else {
				require.NoError(t, err)
				assert.Equal(t, int64(1), submission.ID)
			}
		})
	}
}

func TestSubmissionGetAllForGroup(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()

	t.Run("Admin retrieves all submissions for a group", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
			{ID: 2, TaskID: 2, UserID: 2, Status: types.SubmissionStatusEvaluated},
		}

		setup.submissionRepository.EXPECT().GetAllForGroup(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := setup.service.GetAllForGroup(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher retrieves submissions for their group", func(t *testing.T) {
		expectedGroup := &models.Group{ID: 1, CreatedBy: 2}
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
			{ID: 2, TaskID: 2, UserID: 2, Status: types.SubmissionStatusEvaluated},
		}

		setup.groupRepository.EXPECT().Get(gomock.Any(), int64(1)).Return(expectedGroup, nil).Times(1)
		setup.submissionRepository.EXPECT().GetAllForGroup(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := setup.service.GetAllForGroup(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher tries to retrieve submissions for a group they don't own", func(t *testing.T) {
		expectedGroup := &models.Group{ID: 1, CreatedBy: 3}

		setup.groupRepository.EXPECT().Get(gomock.Any(), int64(1)).Return(expectedGroup, nil).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := setup.service.GetAllForGroup(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
	t.Run("Teacher tries to retrieve submissions for a group but can't get group", func(t *testing.T) {
		setup.groupRepository.EXPECT().Get(gomock.Any(), int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := setup.service.GetAllForGroup(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Student tries to retrieve submissions for a group", func(t *testing.T) {
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := setup.service.GetAllForGroup(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Error retrieving submissions for a group", func(t *testing.T) {
		setup.submissionRepository.EXPECT().GetAllForGroup(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			nil, gorm.ErrInvalidData,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := setup.service.GetAllForGroup(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
}

func TestSubmissionGetAllForUser(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()

	t.Run("Admin retrieves all submissions for a user", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
			{ID: 2, TaskID: 2, UserID: 1, Status: types.SubmissionStatusEvaluated},
		}

		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := setup.service.GetAllForUser(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Student retrieves their own submissions", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
		}

		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := setup.service.GetAllForUser(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 1)
	})

	t.Run("Student tries to retrieve another user's submissions", func(t *testing.T) {
		user := schemas.User{Role: "student", ID: 2}
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
		}
		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}

		submissions, err := setup.service.GetAllForUser(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Teacher retrieves submissions for their tasks", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived, Task: models.Task{CreatedBy: 2}},
			{ID: 2, TaskID: 2, UserID: 1, Status: types.SubmissionStatusEvaluated, Task: models.Task{CreatedBy: 2}},
		}

		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := setup.service.GetAllForUser(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher tries to retrieve submissions for tasks they didn't create", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived, Task: models.Task{CreatedBy: 3}},
		}

		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := setup.service.GetAllForUser(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Empty(t, submissions)
	})

	t.Run("Error retrieving submissions", func(t *testing.T) {
		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			nil, gorm.ErrInvalidData,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := setup.service.GetAllForUser(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
}

func TestGetAllForUserShort(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()

	t.Run("Admin retrieves short submissions for a user", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{
				ID:     1,
				TaskID: 1,
				UserID: 1,
				Result: &models.SubmissionResult{
					TestResults: []models.TestResult{
						{Passed: &trueValue},
						{Passed: &falseValue},
					},
				},
			},
			{
				ID:     2,
				TaskID: 2,
				UserID: 1,
				Result: &models.SubmissionResult{
					TestResults: []models.TestResult{
						{Passed: &trueValue},
						{Passed: &trueValue},
					},
				},
			},
		}

		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := setup.service.GetAllForUserShort(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
		assert.Equal(t, int64(1), submissions[0].ID)
		assert.False(t, submissions[0].Passed)
		assert.Equal(t, int64(1), submissions[0].HowManyPassed)
		assert.Equal(t, int64(2), submissions[1].ID)
		assert.True(t, submissions[1].Passed)
		assert.Equal(t, int64(2), submissions[1].HowManyPassed)
	})

	t.Run("Student tries to retrieve another user's short submissions", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{
				ID:     1,
				TaskID: 1,
				UserID: 1,
				Result: &models.SubmissionResult{
					TestResults: []models.TestResult{
						{Passed: &trueValue},
						{Passed: &falseValue},
					},
				},
			},
			{
				ID:     2,
				TaskID: 2,
				UserID: 1,
				Result: &models.SubmissionResult{
					TestResults: []models.TestResult{
						{Passed: &trueValue},
						{Passed: &trueValue},
					},
				},
			},
		}

		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 2}

		submissions, err := setup.service.GetAllForUserShort(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Error retrieving short submissions", func(t *testing.T) {
		setup.submissionRepository.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			nil, gorm.ErrInvalidData,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := setup.service.GetAllForUserShort(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
}

func TestGetAllForTask(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()

	t.Run("Admin retrieves all submissions for a task", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
			{ID: 2, TaskID: 1, UserID: 2, Status: types.SubmissionStatusEvaluated},
		}

		setup.submissionRepository.EXPECT().GetAllForTask(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := setup.service.GetAllForTask(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher retrieves submissions for their task", func(t *testing.T) {
		expectedTask := &schemas.TaskDetailed{ID: 1, CreatedBy: 2}
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
			{ID: 2, TaskID: 1, UserID: 2, Status: types.SubmissionStatusEvaluated},
		}

		setup.taskService.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(expectedTask, nil).Times(1)
		setup.submissionRepository.EXPECT().GetAllForTask(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := setup.service.GetAllForTask(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher tries to retrieve submissions for a task they didn't create", func(t *testing.T) {
		expectedTask := &schemas.TaskDetailed{ID: 1, CreatedBy: 2}
		setup.taskService.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(expectedTask, nil).Times(1)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 3}

		submissions, err := setup.service.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
	t.Run("Teacher tries to retrieve submissions for a task, but task get fails", func(t *testing.T) {
		setup.taskService.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 3}

		submissions, err := setup.service.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Student retrieves submissions for a task they are assigned to", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: types.SubmissionStatusReceived},
		}

		setup.taskRepository.EXPECT().IsAssignedToUser(gomock.Any(), int64(1), int64(1)).Return(true, nil).Times(1)
		setup.submissionRepository.EXPECT().GetAllForTaskByUser(gomock.Any(), int64(1), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := setup.service.GetAllForTask(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 1)
	})

	t.Run("Student retrieves submissions for a task, but can't check if he is assigned", func(t *testing.T) {
		setup.taskRepository.EXPECT().IsAssignedToUser(gomock.Any(), int64(1), int64(1)).Return(false, gorm.ErrRecordNotFound).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := setup.service.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
	t.Run("Student tries to retrieve submissions for a task they are not assigned to", func(t *testing.T) {
		setup.taskRepository.EXPECT().IsAssignedToUser(gomock.Any(), int64(1), int64(1)).Return(false, nil).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := setup.service.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Error retrieving submissions for a task", func(t *testing.T) {
		setup.submissionRepository.EXPECT().GetAllForTask(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			nil, gorm.ErrInvalidData,
		).Times(1)

		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := setup.service.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
}
func TestMarkSubmissionStatus(t *testing.T) {
	setup := setupSubmissionServiceTest(t)
	defer setup.ctrl.Finish()

	testCases := []struct {
		name          string
		method        func(s service.SubmissionService, tx *database.DB, id int64, args ...interface{}) error
		expectedCall  func()
		expectedError bool
	}{
		{
			name: "Successfully mark submission as failed",
			method: func(s service.SubmissionService, tx *database.DB, id int64, args ...interface{}) error {
				return s.MarkFailed(tx, id, args[0].(string))
			},
			expectedCall: func() {
				setup.submissionRepository.EXPECT().MarkFailed(gomock.Any(), int64(1), "Error message").Return(nil).Times(1)
			},
			expectedError: false,
		},
		{
			name: "Error marking submission as failed",
			method: func(s service.SubmissionService, tx *database.DB, id int64, args ...interface{}) error {
				return s.MarkFailed(tx, id, args[0].(string))
			},
			expectedCall: func() {
				setup.submissionRepository.EXPECT().MarkFailed(gomock.Any(), int64(1), "Error message").Return(gorm.ErrInvalidData).Times(1)
			},
			expectedError: true,
		},
		{
			name: "Successfully mark submission as complete",
			method: func(s service.SubmissionService, tx *database.DB, id int64, _ ...interface{}) error {
				return s.MarkComplete(tx, id)
			},
			expectedCall: func() {
				setup.submissionRepository.EXPECT().MarkEvaluated(gomock.Any(), int64(1)).Return(nil).Times(1)
			},
			expectedError: false,
		},
		{
			name: "Error marking submission as complete",
			method: func(s service.SubmissionService, tx *database.DB, id int64, _ ...interface{}) error {
				return s.MarkComplete(tx, id)
			},
			expectedCall: func() {
				setup.submissionRepository.EXPECT().MarkEvaluated(gomock.Any(), int64(1)).Return(gorm.ErrInvalidData).Times(1)
			},
			expectedError: true,
		},
		{
			name: "Successfully mark submission as processing",
			method: func(s service.SubmissionService, tx *database.DB, id int64, _ ...interface{}) error {
				return s.MarkProcessing(tx, id)
			},
			expectedCall: func() {
				setup.submissionRepository.EXPECT().MarkProcessing(gomock.Any(), int64(1)).Return(nil).Times(1)
			},
			expectedError: false,
		},
		{
			name: "Error marking submission as processing",
			method: func(s service.SubmissionService, tx *database.DB, id int64, _ ...interface{}) error {
				return s.MarkProcessing(tx, id)
			},
			expectedCall: func() {
				setup.submissionRepository.EXPECT().MarkProcessing(gomock.Any(), int64(1)).Return(gorm.ErrInvalidData).Times(1)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.expectedCall()
			var err error
			if tc.name == "Successfully mark submission as failed" || tc.name == "Error marking submission as failed" {
				err = tc.method(setup.service, nil, 1, "Error message")
			} else {
				err = tc.method(setup.service, nil, 1)
			}
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
