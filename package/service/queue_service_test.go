package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// mockQueuePublisher is a simple mock implementation of queue.Publisher
type mockQueuePublisher struct {
	connected     bool
	publishCalled bool
	publishError  error
	body          []byte
}

func (m *mockQueuePublisher) Publish(ctx context.Context, queueName string, replyTo string, body []byte) error {
	m.publishCalled = true
	m.body = body
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
		mock_repository.NewMockContestRepository(ctrl),
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

	db := &testutils.MockDatabase{}
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
		mock_repository.NewMockContestRepository(ctrl),
		submissionRepository,
		submissionResultRepository,
		queueMessageRepository,
		mockQueueClient,
		"test_queue",
		"test_response_queue",
	)
	assert.NotNil(t, queueService)

	db := &testutils.MockDatabase{}

	// Call RetryPendingSubmissions - should return error but not panic
	err := queueService.RetryPendingSubmissions(db)
	require.Error(t, err, "RetryPendingSubmissions should return error when queue is unavailable")
	assert.ErrorIs(t, err, errors.ErrQueueNotConnected)
}

func TestQueueService_PublishSubmission_ContestLimitOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	submissionRepository := mock_repository.NewMockSubmissionRepository(ctrl)
	submissionResultRepository := mock_repository.NewMockSubmissionResultRepository(ctrl)
	contestRepository := mock_repository.NewMockContestRepository(ctrl)
	publisher := &mockQueuePublisher{connected: true}
	queueService := service.NewQueueService(
		mock_repository.NewMockTaskRepository(ctrl),
		contestRepository,
		submissionRepository,
		submissionResultRepository,
		mock_repository.NewMockQueueMessageRepository(ctrl),
		publisher,
		"q",
		"rq",
	)

	contestID := int64(7)
	timeOverride := int64(5000)
	db := &testutils.MockDatabase{}
	submissionRepository.EXPECT().Get(db, int64(1)).Return(&models.Submission{ID: 1, TaskID: 2, ContestID: &contestID}, nil)
	submissionResultRepository.EXPECT().Get(db, int64(1)).Return(&models.SubmissionResult{
		ID:          1,
		TestResults: []models.TestResult{{TestCase: models.TestCase{Order: 1, TimeLimitMs: 1000, MemoryLimitKB: 20000}}},
	}, nil)
	// only the time limit is overridden; memory falls back to the task's value
	contestRepository.EXPECT().GetContestTask(db, contestID, int64(2)).Return(&models.ContestTask{TimeLimitMs: &timeOverride}, nil)
	submissionRepository.EXPECT().MarkProcessing(db, int64(1)).Return(nil)

	require.NoError(t, queueService.PublishSubmission(db, 1, 1))

	var msg schemas.QueueMessage
	require.NoError(t, json.Unmarshal(publisher.body, &msg))
	var payload schemas.TaskQueueMessage
	require.NoError(t, json.Unmarshal(msg.Payload, &payload))
	require.Len(t, payload.TestCases, 1)
	assert.Equal(t, int64(5000), payload.TestCases[0].TimeLimitMs)
	assert.Equal(t, int64(20000), payload.TestCases[0].MemoryLimitKB)
}
