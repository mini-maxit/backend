package service_test

import (
	"testing"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestQueueService_PublishSubmission_WithoutChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepository := mock_repository.NewMockTaskRepository(ctrl)
	submissionRepository := mock_repository.NewMockSubmissionRepository(ctrl)
	submissionResultRepository := mock_repository.NewMockSubmissionResultRepository(ctrl)
	queueMessageRepository := mock_repository.NewMockQueueMessageRepository(ctrl)

	// Create queue service without channel (simulating queue unavailable)
	queueService, err := service.NewQueueService(
		taskRepository,
		submissionRepository,
		submissionResultRepository,
		queueMessageRepository,
		nil, // no connection
		nil, // no channel
		"test_queue",
		"test_response_queue",
	)
	require.NoError(t, err)
	assert.NotNil(t, queueService)

	// Mock submission data
	submission := &models.Submission{
		ID:         1,
		TaskID:     1,
		UserID:     1,
		Order:      1,
		LanguageID: 1,
		Status:     types.SubmissionStatusReceived,
		Language: models.LanguageConfig{
			ID:      1,
			Type:    "cpp",
			Version: "20",
		},
		File: models.File{
			ID:         1,
			Filename:   "solution.cpp",
			Path:       "/path/to/solution.cpp",
			Bucket:     "submissions",
			ServerType: "local",
		},
	}

	submissionResult := &models.SubmissionResult{
		ID:           1,
		SubmissionID: 1,
		Code:         types.SubmissionResultCodeUnknown,
		Message:      "Awaiting processing",
		TestResults:  []models.TestResult{},
	}

	db := &gorm.DB{}
	submissionRepository.EXPECT().Get(db, int64(1)).Return(submission, nil)
	submissionResultRepository.EXPECT().Get(db, int64(1)).Return(submissionResult, nil)

	// Call PublishSubmission - should not fail even without queue channel
	err = queueService.PublishSubmission(db, 1, 1)
	assert.NoError(t, err, "PublishSubmission should not fail when queue is unavailable")
}

func TestQueueService_RetryPendingSubmissions_WithoutChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepository := mock_repository.NewMockTaskRepository(ctrl)
	submissionRepository := mock_repository.NewMockSubmissionRepository(ctrl)
	submissionResultRepository := mock_repository.NewMockSubmissionResultRepository(ctrl)
	queueMessageRepository := mock_repository.NewMockQueueMessageRepository(ctrl)

	// Create queue service without channel (simulating queue unavailable)
	queueService, err := service.NewQueueService(
		taskRepository,
		submissionRepository,
		submissionResultRepository,
		queueMessageRepository,
		nil, // no connection
		nil, // no channel
		"test_queue",
		"test_response_queue",
	)
	require.NoError(t, err)
	assert.NotNil(t, queueService)

	db := &gorm.DB{}

	// Call RetryPendingSubmissions - should return error but not panic
	err = queueService.RetryPendingSubmissions(db)
	require.Error(t, err, "RetryPendingSubmissions should return error when queue is unavailable")
	assert.Contains(t, err.Error(), "queue channel not available")
}
