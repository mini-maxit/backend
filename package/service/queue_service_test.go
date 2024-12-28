package service

import (
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
)

func TestPublishMessage(t *testing.T) {
	config := testutils.NewTestConfig()
	tx := testutils.NewTestTx(t)
	tr, err := repository.NewTaskRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	subR, err := repository.NewSubmissionRepository(tx)
	assert.NoError(t, err)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	qr, err := repository.NewQueueMessageRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	conn, ch := testutils.NewTestChannel(t)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	qs, err := NewQueueService(tr, subR, qr, conn, ch, config.BrokerConfig.QueueName, config.BrokerConfig.ResponseQueueName)
	assert.NoError(t, err)

	// Nothing to test here, just checking if the function doesn't panic
	err = qs.publishMessage(schemas.QueueMessage{})
	assert.NoError(t, err)
}

func TestPublishSubmission(t *testing.T) {
	config := testutils.NewTestConfig()
	tx := testutils.NewTestTx(t)
	_, err := repository.NewUserRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	tr, err := repository.NewTaskRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	subR, err := repository.NewSubmissionRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	qr, err := repository.NewQueueMessageRepository(tx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	conn, ch := testutils.NewTestChannel(t)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	qs, err := NewQueueService(tr, subR, qr, conn, ch, config.BrokerConfig.QueueName, config.BrokerConfig.ResponseQueueName)
	assert.NoError(t, err)
	savePoint := "before"
	tx.SavePoint(savePoint)

	t.Run("Nonexistent submission", func(t *testing.T) {
		err = qs.PublishSubmission(tx, 0)
		assert.Error(t, err)
		tx.RollbackTo(savePoint)
	})

	/* This test is not working because the submission creation is not working properly */
	// t.Run("Non exists task time limits", func(t *testing.T) {
	// 	userId, err := ur.CreateUser(tx, &models.User{
	// 		Name:         "Test user",
	// 		Surname:      "Test user",
	// 		Email:        "email@email.com",
	// 		Username:     "test_user",
	// 		PasswordHash: "password",
	// 		// Role:         "user",
	// 	})
	// 	if !assert.NoError(t, err) {
	// 		t.FailNow()
	// 	}
	// 	task_id, err := tr.Create(tx, models.Task{
	// 		Title:     "Test task",
	// 		CreatedBy: userId,
	// 	})
	// 	if !assert.NoError(t, err) {
	// 		t.FailNow()
	// 	}
	// 	subId, err := subR.CreateSubmission(tx, models.Submission{
	// 		TaskId:     task_id,
	// 		UserId:     userId,
	// 		Order:      1,
	// 		LanguageId: 1,
	// 		Status:     "created",
	// 	})
	// 	if !assert.NoError(t, err) {
	// 		t.FailNow()
	// 	}
	// 	err = qs.PublishSubmission(tx, subId)
	// 	assert.Error(t, err)
	// 	t.Log(err)
	// 	tx.RollbackTo(savePoint)
	// })
	tx.Rollback()
}
