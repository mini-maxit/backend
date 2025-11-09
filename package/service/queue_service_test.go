package service_test

import (
	"context"
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

// mockQueuePublisher is a simple mock implementation of queue.Publisher
type mockQueuePublisher struct {
	connected     bool
	publishCalled bool
	publishError  error
}

func (m *mockQueuePublisher) Publish(ctx context.Context, queueName string, replyTo string, body []byte) error {
	m.publishCalled = true
	return m.publishError
}

func (m *mockQueuePublisher) IsConnected() bool {
	return m.connected
}

func TestQueueService_PublishSubmission_WithoutChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepository := mock_repository.NewMockTaskRepository(ctrl)
	submissionRepository := mock_repository.NewMockSubmissionRepository(ctrl)
	submissionResultRepository := mock_repository.NewMockSubmissionResultRepository(ctrl)
	queueMessageRepository := mock_repository.NewMockQueueMessageRepository(ctrl)

	// Create mock queue client that is not connected
	mockQueueClient := &mockQueuePublisher{
		connected: false,
	}

	// Create queue service
	queueService := service.NewQueueService(
		taskRepository,
		submissionRepository,
		submissionResultRepository,
		queueMessageRepository,
		mockQueueClient,
		"test_queue",
		"test_response_queue",
	)
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

	// Call PublishSubmission - should not fail even without queue connection
	err := queueService.PublishSubmission(db, 1, 1)
	require.NoError(t, err, "PublishSubmission should not fail when queue is unavailable")
	assert.False(t, mockQueueClient.publishCalled, "Publish should not be called when queue is not connected")
}

func TestQueueService_RetryPendingSubmissions_WithoutChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepository := mock_repository.NewMockTaskRepository(ctrl)
	submissionRepository := mock_repository.NewMockSubmissionRepository(ctrl)
	submissionResultRepository := mock_repository.NewMockSubmissionResultRepository(ctrl)
	queueMessageRepository := mock_repository.NewMockQueueMessageRepository(ctrl)

	// Create mock queue client that is not connected
	mockQueueClient := &mockQueuePublisher{
		connected: false,
	}

	// Create queue service
	queueService := service.NewQueueService(
		taskRepository,
		submissionRepository,
		submissionResultRepository,
		queueMessageRepository,
		mockQueueClient,
		"test_queue",
		"test_response_queue",
	)
	assert.NotNil(t, queueService)

	db := &gorm.DB{}

	// Call RetryPendingSubmissions - should return error but not panic
	err := queueService.RetryPendingSubmissions(db)
	require.Error(t, err, "RetryPendingSubmissions should return error when queue is unavailable")
	assert.Contains(t, err.Error(), "queue not connected")
}
