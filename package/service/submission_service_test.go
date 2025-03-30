package service

import (
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := testutils.NewMockSubmissionRepository(ctrl)

	submission := &models.Submission{
		TaskId: 1,
		UserId: 1,

		Order:      1,
		LanguageId: 1,
		Status:     string(models.StatusReceived),
	}

	t.Run("Succesful", func(t *testing.T) {
		submissionId := rand.Int64()
		expectedModel := &models.Submission{
			TaskId:     rand.Int64(),
			UserId:     rand.Int64(),
			Order:      rand.Int64(),
			LanguageId: rand.Int64(),
			Status:     models.StatusReceived,
		}
		m.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&models.Submission{})).DoAndReturn(func(tx *gorm.DB, s *models.Submission) (int64, error) {
			if !reflect.DeepEqual(s, expectedModel) {
				t.Fatalf("invalid submission model passed to repo. exp=%v got=%v", expectedModel, s)
			}
			submission.Id = submissionId
			submission.SubmittedAt = time.Now()
			return submission.Id, nil
		}).Times(1)
		s := NewSubmissionService(m, nil, nil, nil, nil, nil, nil, nil, nil)

		id, err := s.Create(nil, expectedModel.TaskId, expectedModel.UserId, expectedModel.LanguageId, expectedModel.Order)
		assert.NoError(t, err)
		assert.Equal(t, submissionId, id)
	})

	t.Run("Fails if repository fails unexpectedly", func(t *testing.T) {
		m.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&models.Submission{})).DoAndReturn(func(tx *gorm.DB, s *models.Submission) (int64, error) {
			return 0, gorm.ErrInvalidData
		}).Times(1)
		s := NewSubmissionService(m, nil, nil, nil, nil, nil, nil, nil, nil)
		id, err := s.Create(nil, 1, 1, 1, 1)
		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
	})
}
