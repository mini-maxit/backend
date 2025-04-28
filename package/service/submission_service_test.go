package service_test

import (
	"encoding/json"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := testutils.NewMockSubmissionRepository(ctrl)

	submission := &models.Submission{
		TaskID: 1,
		UserID: 1,

		Order:      1,
		LanguageID: 1,
		Status:     models.StatusReceived,
	}

	t.Run("Succesful", func(t *testing.T) {
		submissionID := rand.Int64()
		expectedModel := &models.Submission{
			TaskID:     rand.Int64(),
			UserID:     rand.Int64(),
			Order:      rand.Int64(),
			LanguageID: rand.Int64(),
			Status:     models.StatusReceived,
		}
		m.EXPECT().Create(
			gomock.Any(),
			gomock.AssignableToTypeOf(&models.Submission{}),
		).DoAndReturn(func(_ *gorm.DB, s *models.Submission) (int64, error) {
			if !reflect.DeepEqual(s, expectedModel) {
				t.Fatalf("invalid submission model passed to repo. exp=%v got=%v", expectedModel, s)
			}
			submission.ID = submissionID
			submission.SubmittedAt = time.Now()
			return submission.ID, nil
		}).Times(1)
		s := service.NewSubmissionService(m, nil, nil, nil, nil, nil, nil, nil, nil)

		id, err := s.Create(nil, expectedModel.TaskID, expectedModel.UserID, expectedModel.LanguageID, expectedModel.Order)
		require.NoError(t, err)
		assert.Equal(t, submissionID, id)
	})

	t.Run("Fails if repository fails unexpectedly", func(t *testing.T) {
		m.EXPECT().
			Create(gomock.Any(), gomock.AssignableToTypeOf(&models.Submission{})).
			DoAndReturn(func(_ *gorm.DB, _ *models.Submission) (int64, error) {
				return 0, gorm.ErrInvalidData
			}).
			Times(1)
		s := service.NewSubmissionService(m, nil, nil, nil, nil, nil, nil, nil, nil)
		id, err := s.Create(nil, 1, 1, 1, 1)
		require.Error(t, err)
		assert.Equal(t, int64(0), id)
	})
}

func TestCreateSubmissionResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mSubmissionRepo := testutils.NewMockSubmissionRepository(ctrl)
	mSubmissionResRepo := mock_repository.NewMockSubmissionResultRepository(ctrl)
	mIORepo := mock_repository.NewMockInputOutputRepository(ctrl)
	mTestRepo := mock_repository.NewMockTestRepository(ctrl)
	submission := &models.Submission{
		ID:          1,
		TaskID:      1,
		UserID:      1,
		Order:       1,
		LanguageID:  1,
		Status:      models.StatusReceived,
		SubmittedAt: time.Now(),
	}

	submissionResultID := rand.Int64()
	testCases := []struct {
		name          string
		setupMocks    func()
		expectedID    int64
		expectedErr   bool
		queueResponse schemas.QueueResponseMessage
	}{
		{
			name: "Could not get submission",
			setupMocks: func() {
				mSubmissionRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)
			},
			expectedID:    int64(-1),
			expectedErr:   true,
			queueResponse: schemas.QueueResponseMessage{},
		},
		{
			name: "Could not unmarshal response message payload",
			setupMocks: func() {
				mSubmissionRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(submission, nil).Times(1)
			},
			expectedID:    int64(-1),
			expectedErr:   true,
			queueResponse: schemas.QueueResponseMessage{},
		},
		{
			name: "Could not create submission result",
			setupMocks: func() {
				mSubmissionRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(submission, nil).Times(1)
				mSubmissionResRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(0), gorm.ErrInvalidData).Times(1)
			},
			expectedID:  int64(-1),
			expectedErr: true,
			queueResponse: schemas.QueueResponseMessage{Payload: json.RawMessage(`
				{
					"type": "handshake",
					"message_id": "adsa",
					"ok": true,
					"payload": {
					    "languages": [
					    {
						    "name": "CPP",
					        "versions": ["20", "17"]
					    },
					    {
					        "name": "Python",
					        "versions": ["3", "2"]
					    }
					    ]
				  }
				}`)},
		},
		{
			name: "Could not get input output",
			setupMocks: func() {
				mSubmissionRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(submission, nil).Times(1)
				mSubmissionResRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(1), nil).Times(1)
				mIORepo.EXPECT().GetInputOutputID(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					int64(0), gorm.ErrRecordNotFound,
				).Times(1)
			},
			expectedID:  int64(-1),
			expectedErr: true,
			queueResponse: schemas.QueueResponseMessage{Payload: json.RawMessage(`
				{"success":true,
				"statuscode":1,
				"message":"solution executed successfully",
				"testresults":
				[
					{"passed":false,"errormessage":"difference at line 1:\noutput:   hello, world!\nexpected: hello world!\n\n",
					"order":1},
					{"passed":true,"errormessage":"","order":2}]}
			`)},
		},
		{
			name: "Could not create test result",
			setupMocks: func() {
				mSubmissionRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(submission, nil).Times(1)
				mSubmissionResRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(1), nil).Times(1)
				mIORepo.EXPECT().GetInputOutputID(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).DoAndReturn(func(_ *gorm.DB, _ int64, order int) (int64, error) {
					return int64(order), nil
				}).Times(1)
				mTestRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidData).Times(1)
			},
			expectedID:  int64(-1),
			expectedErr: true,
			queueResponse: schemas.QueueResponseMessage{Payload: json.RawMessage(`
				{"success":true,
				"statuscode":1,
				"message":"solution executed successfully",
				"testresults":
				[
					{"passed":false,"errormessage":"difference at line 1:\noutput:   hello, world!\nexpected: hello world!\n\n",
					"order":1},
					{"passed":true,"errormessage":"","order":2}]}
			`)},
		},
		{
			name: "Success",
			setupMocks: func() {
				mSubmissionRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(submission, nil).Times(1)
				mSubmissionResRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(submissionResultID, nil).Times(1)
				mIORepo.EXPECT().GetInputOutputID(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).DoAndReturn(func(_ *gorm.DB, _ int64, order int) (int64, error) {
					return int64(order), nil
				}).Times(2)
				mTestRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).Times(2)
			},
			expectedID:  submissionResultID,
			expectedErr: false,
			queueResponse: schemas.QueueResponseMessage{Payload: json.RawMessage(`
				{"success":true,
				"statuscode":1,
				"message":"solution executed successfully",
				"testresults":
				[
					{"passed":false,"errormessage":"difference at line 1:\noutput:   hello, world!\nexpected: hello world!\n\n",
					"order":1},
					{"passed":true,"errormessage":"","order":2}]}
			`)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			s := service.NewSubmissionService(mSubmissionRepo, mSubmissionResRepo, mIORepo, mTestRepo, nil, nil, nil, nil, nil)
			id, err := s.CreateSubmissionResult(nil, submission.ID, tc.queueResponse)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLanguageService := mock_service.NewMockLanguageService(ctrl)

	t.Run("Success", func(t *testing.T) {
		expectedLanguages := []schemas.LanguageConfig{
			{Type: "Python", Version: "3.8"},
			{Type: "Python", Version: "3.9"},
			{Type: "Go", Version: "1.22"},
			{Type: "Go", Version: "1.23"},
		}

		mLanguageService.EXPECT().GetAllEnabled(gomock.Any()).Return(expectedLanguages, nil).Times(1)

		s := service.NewSubmissionService(nil, nil, nil, nil, nil, nil, mLanguageService, nil, nil)
		languages, err := s.GetAvailableLanguages(nil)

		require.NoError(t, err)
		assert.Equal(t, expectedLanguages, languages)
	})

	t.Run("Error fetching languages", func(t *testing.T) {
		mLanguageService.EXPECT().GetAllEnabled(gomock.Any()).Return(nil, gorm.ErrRecordNotFound).Times(1)

		s := service.NewSubmissionService(nil, nil, nil, nil, nil, nil, mLanguageService, nil, nil)
		languages, err := s.GetAvailableLanguages(nil)

		require.Error(t, err)
		assert.Nil(t, languages)
	})
}
func TestGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mSubmissionRepo := mock_repository.NewMockSubmissionRepository(ctrl)

	testCases := []struct {
		name           string
		user           schemas.User
		expectedMethod func() *gomock.Call
		expectedResult []schemas.Submission
		expectedErr    bool
	}{
		{
			name: "Admin retrieves all submissions",
			user: schemas.User{Role: "admin"},
			expectedMethod: func() *gomock.Call {
				return mSubmissionRepo.EXPECT().GetAll(gomock.Any(), 10, 0, "submitted_at:desc").Return([]models.Submission{
					{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
					{ID: 2, TaskID: 2, UserID: 2, Status: models.StatusEvaluated},
				}, nil).Times(1)
			},
			expectedResult: []schemas.Submission{
				{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
				{ID: 2, TaskID: 2, UserID: 2, Status: models.StatusEvaluated},
			},
			expectedErr: false,
		},
		{
			name: "Teacher retrieves submissions for their tasks",
			user: schemas.User{Role: "teacher", ID: 1},
			expectedMethod: func() *gomock.Call {
				return mSubmissionRepo.EXPECT().GetAllForTeacher(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
					[]models.Submission{
						{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
						{ID: 2, TaskID: 2, UserID: 2, Status: models.StatusEvaluated},
					}, nil).Times(1)
			},
			expectedResult: []schemas.Submission{
				{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
				{ID: 2, TaskID: 2, UserID: 2, Status: models.StatusEvaluated},
			},
			expectedErr: false,
		},
		{
			name: "Student retrieves their own submissions",
			user: schemas.User{Role: "student", ID: 1},
			expectedMethod: func() *gomock.Call {
				return mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
					[]models.Submission{
						{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
					}, nil).Times(1)
			},
			expectedResult: []schemas.Submission{
				{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
			},
			expectedErr: false,
		},
		{
			name: "Error retrieving submissions",
			user: schemas.User{Role: "admin"},
			expectedMethod: func() *gomock.Call {
				return mSubmissionRepo.EXPECT().GetAll(gomock.Any(), 10, 0, "submitted_at:desc").Return(
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
			s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
			queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}

			submissions, err := s.GetAll(nil, tc.user, queryParams)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mSubmissionRepo := mock_repository.NewMockSubmissionRepository(ctrl)

	testCases := []struct {
		name               string
		user               schemas.User
		expectedSubmission *models.Submission
		expectedErr        bool
	}{
		{
			name:               "Admin retrieves a submission",
			user:               schemas.User{Role: "admin"},
			expectedSubmission: &models.Submission{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
			expectedErr:        false,
		},
		{
			name:               "Student tries to access another user's submission",
			user:               schemas.User{Role: "student", ID: 1},
			expectedSubmission: &models.Submission{ID: 1, TaskID: 1, UserID: 2, Status: models.StatusReceived},
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
				mSubmissionRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(tc.expectedSubmission, nil).Times(1)
			} else {
				mSubmissionRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)
			}

			s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
			submission, err := s.Get(nil, 1, tc.user)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mSubmissionRepo := mock_repository.NewMockSubmissionRepository(ctrl)
	mGroupRepo := mock_repository.NewMockGroupRepository(ctrl)

	t.Run("Admin retrieves all submissions for a group", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
			{ID: 2, TaskID: 2, UserID: 2, Status: models.StatusEvaluated},
		}

		mSubmissionRepo.EXPECT().GetAllForGroup(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, mGroupRepo, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := s.GetAllForGroup(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher retrieves submissions for their group", func(t *testing.T) {
		expectedGroup := &models.Group{ID: 1, CreatedBy: 2}
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
			{ID: 2, TaskID: 2, UserID: 2, Status: models.StatusEvaluated},
		}

		mGroupRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(expectedGroup, nil).Times(1)
		mSubmissionRepo.EXPECT().GetAllForGroup(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, mGroupRepo, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := s.GetAllForGroup(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher tries to retrieve submissions for a group they don't own", func(t *testing.T) {
		expectedGroup := &models.Group{ID: 1, CreatedBy: 3}

		mGroupRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(expectedGroup, nil).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, mGroupRepo, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := s.GetAllForGroup(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
	t.Run("Teacher tries to retrieve submissions for a group but can't get group", func(t *testing.T) {
		mGroupRepo.EXPECT().Get(gomock.Any(), int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, mGroupRepo, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := s.GetAllForGroup(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Student tries to retrieve submissions for a group", func(t *testing.T) {
		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, mGroupRepo, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := s.GetAllForGroup(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Error retrieving submissions for a group", func(t *testing.T) {
		mSubmissionRepo.EXPECT().GetAllForGroup(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			nil, gorm.ErrInvalidData,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, mGroupRepo, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := s.GetAllForGroup(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
}

func TestSubmissionGetAllForUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mSubmissionRepo := mock_repository.NewMockSubmissionRepository(ctrl)

	t.Run("Admin retrieves all submissions for a user", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
			{ID: 2, TaskID: 2, UserID: 1, Status: models.StatusEvaluated},
		}

		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := s.GetAllForUser(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Student retrieves their own submissions", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
		}

		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := s.GetAllForUser(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 1)
	})

	t.Run("Student tries to retrieve another user's submissions", func(t *testing.T) {
		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)

		user := schemas.User{Role: "student", ID: 2}
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
		}
		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}

		submissions, err := s.GetAllForUser(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Teacher retrieves submissions for their tasks", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived, Task: models.Task{CreatedBy: 2}},
			{ID: 2, TaskID: 2, UserID: 1, Status: models.StatusEvaluated, Task: models.Task{CreatedBy: 2}},
		}

		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := s.GetAllForUser(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher tries to retrieve submissions for tasks they didn't create", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived, Task: models.Task{CreatedBy: 3}},
		}

		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := s.GetAllForUser(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Empty(t, submissions)
	})

	t.Run("Error retrieving submissions", func(t *testing.T) {
		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			nil, gorm.ErrInvalidData,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := s.GetAllForUser(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
}

func TestGetAllForUserShort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mSubmissionRepo := mock_repository.NewMockSubmissionRepository(ctrl)

	t.Run("Admin retrieves short submissions for a user", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{
				ID:     1,
				TaskID: 1,
				UserID: 1,
				Result: &models.SubmissionResult{
					TestResult: []models.TestResult{
						{Passed: true},
						{Passed: false},
					},
				},
			},
			{
				ID:     2,
				TaskID: 2,
				UserID: 1,
				Result: &models.SubmissionResult{
					TestResult: []models.TestResult{
						{Passed: true},
						{Passed: true},
					},
				},
			},
		}

		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := s.GetAllForUserShort(nil, 1, user, queryParams)

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
					TestResult: []models.TestResult{
						{Passed: true},
						{Passed: false},
					},
				},
			},
			{
				ID:     2,
				TaskID: 2,
				UserID: 1,
				Result: &models.SubmissionResult{
					TestResult: []models.TestResult{
						{Passed: true},
						{Passed: true},
					},
				},
			},
		}

		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)
		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 2}

		submissions, err := s.GetAllForUserShort(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Error retrieving short submissions", func(t *testing.T) {
		mSubmissionRepo.EXPECT().GetAllByUser(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			nil, gorm.ErrInvalidData,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := s.GetAllForUserShort(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
}

func TestGetAllForTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mSubmissionRepo := mock_repository.NewMockSubmissionRepository(ctrl)
	mTaskRepo := mock_repository.NewMockTaskRepository(ctrl)
	mTaskService := mock_service.NewMockTaskService(ctrl)

	t.Run("Admin retrieves all submissions for a task", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
			{ID: 2, TaskID: 1, UserID: 2, Status: models.StatusEvaluated},
		}

		mSubmissionRepo.EXPECT().GetAllForTask(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, mTaskRepo, nil, mTaskService, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := s.GetAllForTask(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher retrieves submissions for their task", func(t *testing.T) {
		expectedTask := &schemas.TaskDetailed{ID: 1, CreatedBy: 2}
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
			{ID: 2, TaskID: 1, UserID: 2, Status: models.StatusEvaluated},
		}

		mTaskService.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(expectedTask, nil).Times(1)
		mSubmissionRepo.EXPECT().GetAllForTask(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, mTaskRepo, nil, mTaskService, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 2}

		submissions, err := s.GetAllForTask(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 2)
	})

	t.Run("Teacher tries to retrieve submissions for a task they didn't create", func(t *testing.T) {
		expectedTask := &schemas.TaskDetailed{ID: 1, CreatedBy: 2}
		mTaskService.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(expectedTask, nil).Times(1)
		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, mTaskRepo, nil, mTaskService, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 3}

		submissions, err := s.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
	t.Run("Teacher tries to retrieve submissions for a task, but task get fails", func(t *testing.T) {
		mTaskService.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(nil, gorm.ErrRecordNotFound).Times(1)
		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, mTaskRepo, nil, mTaskService, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "teacher", ID: 3}

		submissions, err := s.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Student retrieves submissions for a task they are assigned to", func(t *testing.T) {
		expectedSubmissions := []models.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: models.StatusReceived},
		}

		mTaskRepo.EXPECT().IsAssignedToUser(gomock.Any(), int64(1), int64(1)).Return(true, nil).Times(1)
		mSubmissionRepo.EXPECT().GetAllForTaskByUser(gomock.Any(), int64(1), int64(1), 10, 0, "submitted_at:desc").Return(
			expectedSubmissions, nil,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, mTaskRepo, nil, mTaskService, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := s.GetAllForTask(nil, 1, user, queryParams)

		require.NoError(t, err)
		assert.Len(t, submissions, 1)
	})

	t.Run("Student retrieves submissions for a task, but can't check if he is assigned", func(t *testing.T) {
		mTaskRepo.EXPECT().IsAssignedToUser(gomock.Any(), int64(1), int64(1)).Return(false, gorm.ErrRecordNotFound).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, mTaskRepo, nil, mTaskService, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := s.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
	t.Run("Student tries to retrieve submissions for a task they are not assigned to", func(t *testing.T) {
		mTaskRepo.EXPECT().IsAssignedToUser(gomock.Any(), int64(1), int64(1)).Return(false, nil).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, mTaskRepo, nil, mTaskService, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "student", ID: 1}

		submissions, err := s.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})

	t.Run("Error retrieving submissions for a task", func(t *testing.T) {
		mSubmissionRepo.EXPECT().GetAllForTask(gomock.Any(), int64(1), 10, 0, "submitted_at:desc").Return(
			nil, gorm.ErrInvalidData,
		).Times(1)

		s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, mTaskRepo, nil, mTaskService, nil)
		queryParams := map[string]any{"limit": 10, "offset": 0, "sort": "submitted_at:desc"}
		user := schemas.User{Role: "admin"}

		submissions, err := s.GetAllForTask(nil, 1, user, queryParams)

		require.Error(t, err)
		assert.Nil(t, submissions)
	})
}
func TestMarkSubmissionStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mSubmissionRepo := mock_repository.NewMockSubmissionRepository(ctrl)

	testCases := []struct {
		name          string
		method        func(s service.SubmissionService, tx *gorm.DB, id int64, args ...interface{}) error
		expectedCall  func()
		expectedError bool
	}{
		{
			name: "Successfully mark submission as failed",
			method: func(s service.SubmissionService, tx *gorm.DB, id int64, args ...interface{}) error {
				return s.MarkFailed(tx, id, args[0].(string))
			},
			expectedCall: func() {
				mSubmissionRepo.EXPECT().MarkFailed(gomock.Any(), int64(1), "Error message").Return(nil).Times(1)
			},
			expectedError: false,
		},
		{
			name: "Error marking submission as failed",
			method: func(s service.SubmissionService, tx *gorm.DB, id int64, args ...interface{}) error {
				return s.MarkFailed(tx, id, args[0].(string))
			},
			expectedCall: func() {
				mSubmissionRepo.EXPECT().MarkFailed(gomock.Any(), int64(1), "Error message").Return(gorm.ErrInvalidData).Times(1)
			},
			expectedError: true,
		},
		{
			name: "Successfully mark submission as complete",
			method: func(s service.SubmissionService, tx *gorm.DB, id int64, _ ...interface{}) error {
				return s.MarkComplete(tx, id)
			},
			expectedCall: func() {
				mSubmissionRepo.EXPECT().MarkComplete(gomock.Any(), int64(1)).Return(nil).Times(1)
			},
			expectedError: false,
		},
		{
			name: "Error marking submission as complete",
			method: func(s service.SubmissionService, tx *gorm.DB, id int64, _ ...interface{}) error {
				return s.MarkComplete(tx, id)
			},
			expectedCall: func() {
				mSubmissionRepo.EXPECT().MarkComplete(gomock.Any(), int64(1)).Return(gorm.ErrInvalidData).Times(1)
			},
			expectedError: true,
		},
		{
			name: "Successfully mark submission as processing",
			method: func(s service.SubmissionService, tx *gorm.DB, id int64, _ ...interface{}) error {
				return s.MarkProcessing(tx, id)
			},
			expectedCall: func() {
				mSubmissionRepo.EXPECT().MarkProcessing(gomock.Any(), int64(1)).Return(nil).Times(1)
			},
			expectedError: false,
		},
		{
			name: "Error marking submission as processing",
			method: func(s service.SubmissionService, tx *gorm.DB, id int64, _ ...interface{}) error {
				return s.MarkProcessing(tx, id)
			},
			expectedCall: func() {
				mSubmissionRepo.EXPECT().MarkProcessing(gomock.Any(), int64(1)).Return(gorm.ErrInvalidData).Times(1)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.expectedCall()
			s := service.NewSubmissionService(mSubmissionRepo, nil, nil, nil, nil, nil, nil, nil, nil)
			var err error
			if tc.name == "Successfully mark submission as failed" || tc.name == "Error marking submission as failed" {
				err = tc.method(s, nil, 1, "Error message")
			} else {
				err = tc.method(s, nil, 1)
			}
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
