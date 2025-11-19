package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestContestWithStatsToSchema(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	visible := true

	contestWithStats := &models.ContestWithStats{
		Contest: models.Contest{
			ID:          1,
			Name:        "Test Contest",
			Description: "Test Description",
			StartAt:     &startTime,
			EndAt:       &endTime,
			CreatedBy:   1,
			IsVisible:   &visible,
		},
		TaskCount:        5,
		ParticipantCount: 10,
		IsParticipant:    true,
		HasPendingReg:    false,
	}

	result := service.ContestWithStatsToAvailableContest(contestWithStats)

	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Test Contest", result.Name)
	assert.Equal(t, "Test Description", result.Description)
	assert.Equal(t, &startTime, result.StartAt)
	assert.Equal(t, &endTime, result.EndAt)
	assert.Equal(t, int64(1), result.CreatedBy)
	assert.Equal(t, int64(5), result.TaskCount)
	assert.Equal(t, int64(10), result.ParticipantCount)
}

func TestContestWithStatsToSchemaWithNilUserInfo(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	visible := true

	contestWithStats := &models.ContestWithStats{
		Contest: models.Contest{
			ID:          1,
			Name:        "Test Contest",
			Description: "Test Description",
			StartAt:     &startTime,
			EndAt:       &endTime,
			CreatedBy:   1,
			IsVisible:   &visible,
		},
		TaskCount:        5,
		ParticipantCount: 10,
		IsParticipant:    false,
		HasPendingReg:    false,
	}

	result := service.ContestWithStatsToAvailableContest(contestWithStats)

	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Test Contest", result.Name)
	assert.Equal(t, "Test Description", result.Description)
	assert.Equal(t, &startTime, result.StartAt)
	assert.Equal(t, &endTime, result.EndAt)
	assert.Equal(t, int64(1), result.CreatedBy)
	assert.Equal(t, int64(5), result.TaskCount)
	assert.Equal(t, int64(10), result.ParticipantCount)
}

func TestContestWithStatsToSchemaWithMultipleContests(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	visible := true

	contests := []models.ContestWithStats{
		{
			Contest: models.Contest{
				ID:          1,
				Name:        "Contest 1",
				Description: "Description 1",
				StartAt:     &startTime,
				EndAt:       &endTime,
				CreatedBy:   1,
				IsVisible:   &visible,
			},
			TaskCount:        3,
			ParticipantCount: 5,
			IsParticipant:    true,
			HasPendingReg:    false,
		},
		{
			Contest: models.Contest{
				ID:          2,
				Name:        "Contest 2",
				Description: "Description 2",
				StartAt:     &startTime,
				EndAt:       &endTime,
				CreatedBy:   2,
				IsVisible:   &visible,
			},
			TaskCount:        4,
			ParticipantCount: 8,
			IsParticipant:    false,
			HasPendingReg:    false,
		},
	}

	results := make([]*schemas.AvailableContest, len(contests))
	for i, contest := range contests {
		results[i] = service.ContestWithStatsToAvailableContest(&contest)
	}

	assert.Len(t, results, 2)

	// Check first contest
	assert.Equal(t, int64(1), results[0].ID)
	assert.Equal(t, "Contest 1", results[0].Name)
	assert.Equal(t, int64(3), results[0].TaskCount)

	// Check second contest
	assert.Equal(t, int64(2), results[1].ID)
	assert.Equal(t, "Contest 2", results[1].Name)
	assert.Equal(t, int64(4), results[1].TaskCount)
}

func TestContestWithStatsToSchemaWithNilIsVisible(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)

	contestWithStats := &models.ContestWithStats{
		Contest: models.Contest{
			ID:          1,
			Name:        "Test Contest",
			Description: "Test Description",
			StartAt:     &startTime,
			EndAt:       &endTime,
			CreatedBy:   1,
			IsVisible:   nil,
		},
		TaskCount:        5,
		ParticipantCount: 10,
		IsParticipant:    true,
		HasPendingReg:    false,
	}

	result := service.ContestWithStatsToAvailableContest(contestWithStats)

	assert.NotNil(t, result)
}

func TestContestService_GetPastContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	collab := mock_repository.NewMockCollaboratorRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, collab, ts)
	tx := &gorm.DB{}

	t.Run("successful retrieval", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		queryParams := schemas.PaginationParams{
			Limit:  10,
			Offset: 0,
			Sort:   "start_time",
		}

		visible := true
		contestsWithStats := []models.ContestWithStats{
			{
				Contest: models.Contest{
					ID:        1,
					Name:      "Past Contest",
					IsVisible: &visible,
				},
				TaskCount:        5,
				ParticipantCount: 10,
			},
		}

		cr.EXPECT().GetPastContestsWithStats(tx, currentUser.ID, 0, 10, "start_time").Return(contestsWithStats, int64(1), nil).Times(1)

		result, err := cs.GetPastContests(tx, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, int64(1), result.Items[0].ID)
		assert.Equal(t, 1, result.Pagination.TotalItems)
		assert.Equal(t, 1, result.Pagination.CurrentPage)
		assert.Equal(t, 10, result.Pagination.PageSize)
	})

	t.Run("repository error", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		queryParams := schemas.PaginationParams{
			Limit:  10,
			Offset: 0,
			Sort:   "start_time",
		}

		cr.EXPECT().GetPastContestsWithStats(tx, currentUser.ID, 0, 10, "start_time").Return(nil, int64(0), errors.New("db error")).Times(1)

		result, err := cs.GetPastContests(tx, currentUser, queryParams)

		require.Error(t, err)
		assert.Equal(t, 0, result.Pagination.TotalItems)
		assert.Empty(t, result.Items)
	})
}

func TestContestService_GetUpcomingContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	collab := mock_repository.NewMockCollaboratorRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, collab, ts)
	tx := &gorm.DB{}

	t.Run("successful retrieval", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		queryParams := schemas.PaginationParams{
			Limit:  10,
			Offset: 0,
			Sort:   "start_time",
		}

		visible := true
		contestsWithStats := []models.ContestWithStats{
			{
				Contest: models.Contest{
					ID:        1,
					Name:      "Upcoming Contest",
					IsVisible: &visible,
				},
				TaskCount:        5,
				ParticipantCount: 10,
			},
		}

		cr.EXPECT().GetUpcomingContestsWithStats(tx, currentUser.ID, 0, 10, "start_time").Return(contestsWithStats, int64(1), nil).Times(1)

		result, err := cs.GetUpcomingContests(tx, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, int64(1), result.Items[0].ID)
		assert.Equal(t, 1, result.Pagination.TotalItems)
		assert.Equal(t, 1, result.Pagination.CurrentPage)
		assert.Equal(t, 10, result.Pagination.PageSize)
	})

	t.Run("repository error", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		queryParams := schemas.PaginationParams{
			Limit:  10,
			Offset: 0,
			Sort:   "start_time",
		}

		cr.EXPECT().GetUpcomingContestsWithStats(tx, currentUser.ID, 0, 10, "start_time").Return(nil, int64(0), errors.New("db error")).Times(1)

		result, err := cs.GetUpcomingContests(tx, currentUser, queryParams)

		require.Error(t, err)
		assert.Empty(t, result.Items)
		assert.Equal(t, 0, result.Pagination.TotalItems)
	})
}

func TestContestService_ApproveRegistrationRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	collab := mock_repository.NewMockCollaboratorRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, collab, ts)
	tx := &gorm.DB{}

	t.Run("successful approval by admin", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusApproved).Return(nil).Times(1)
		cr.EXPECT().CreateContestParticipant(tx, contestID, userID).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.NoError(t, err)
	})

	t.Run("successful approval by contest creator teacher", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   2,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2, // Same as current user
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusApproved).Return(nil).Times(1)
		cr.EXPECT().CreateContestParticipant(tx, contestID, userID).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.NoError(t, err)
	})

	t.Run("unauthorized - student role", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		userID := int64(5)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrForbidden)
	})

	t.Run("contest not found", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(999)
		userID := int64(5)

		cr.EXPECT().Get(tx, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrNotFound)
	})

	t.Run("teacher not authorized - not contest creator", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   3,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2, // Different from current user
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		collab.EXPECT().HasContestPermission(tx, contestID, currentUser.ID, types.PermissionManage).Return(false, nil).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrForbidden)
	})

	t.Run("user not found", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(999)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrNotFound)
	})

	t.Run("user already participant", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(true, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().DeleteRegistrationRequest(tx, request.ID).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrAlreadyParticipant)
	})

	t.Run("no pending registration", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(nil, nil).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrNoPendingRegistration)
	})

	t.Run("error checking participant status", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, gorm.ErrInvalidDB).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})

	t.Run("error checking pending registration", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(nil, gorm.ErrInvalidDB).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})

	t.Run("error updating registration status", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusApproved).Return(gorm.ErrInvalidDB).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})

	t.Run("error adding user as participant", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusApproved).Return(nil).Times(1)
		cr.EXPECT().CreateContestParticipant(tx, contestID, userID).Return(gorm.ErrInvalidDB).Times(1)

		err := cs.ApproveRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})
}

func TestContestService_RejectRegistrationRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	collab := mock_repository.NewMockCollaboratorRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, collab, ts)
	tx := &gorm.DB{}

	t.Run("successful rejection by admin", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusRejected).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.NoError(t, err)
	})

	t.Run("successful rejection by contest creator teacher", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   2,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusRejected).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.NoError(t, err)
	})

	t.Run("unauthorized - student role", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		userID := int64(5)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrForbidden)
	})

	t.Run("contest not found", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(999)
		userID := int64(5)

		cr.EXPECT().Get(tx, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrNotFound)
	})

	t.Run("teacher not authorized - not contest creator", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   3,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		collab.EXPECT().HasContestPermission(tx, contestID, currentUser.ID, types.PermissionManage).Return(false, nil).Times(1)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrForbidden)
	})

	t.Run("user not found", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(999)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrNotFound)
	})

	t.Run("user already participant", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(true, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().DeleteRegistrationRequest(tx, request.ID).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrAlreadyParticipant)
	})

	t.Run("no pending registration", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(nil, nil).Times(1)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, myerrors.ErrNoPendingRegistration)
	})

	t.Run("error updating registration status", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		userID := int64(5)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
		}

		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		request := &models.ContestRegistrationRequests{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Status:    types.RegistrationRequestStatusPending,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(tx, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusRejected).Return(gorm.ErrInvalidDB).Times(1)

		err := cs.RejectRegistrationRequest(tx, currentUser, contestID, userID)

		require.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})
}

func TestContestService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	collab := mock_repository.NewMockCollaboratorRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, collab, ts)
	tx := &gorm.DB{}

	t.Run("successful retrieval - visible contest", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		visible := true

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: &visible,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)

		result, err := cs.Get(tx, currentUser, contestID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, contestID, result.ID)
		assert.Equal(t, "Test Contest", result.Name)
	})

	t.Run("successful retrieval - invisible contest by admin", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		visible := false

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: &visible,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)

		result, err := cs.Get(tx, currentUser, contestID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, contestID, result.ID)
	})

	t.Run("successful retrieval - invisible contest by creator teacher", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   2,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		visible := false

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: &visible,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)

		result, err := cs.Get(tx, currentUser, contestID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, contestID, result.ID)
	})

	t.Run("unauthorized - student viewing invisible contest", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		visible := false

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: &visible,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)

		result, err := cs.Get(tx, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		require.ErrorIs(t, err, myerrors.ErrForbidden)
	})

	t.Run("unauthorized - teacher viewing invisible contest not created by them", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   3,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		visible := false

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: &visible,
		}

		cr.EXPECT().Get(tx, contestID).Return(contest, nil).Times(1)

		result, err := cs.Get(tx, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		require.ErrorIs(t, err, myerrors.ErrForbidden)
	})

	t.Run("contest not found", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(999)

		cr.EXPECT().Get(tx, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		result, err := cs.Get(tx, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		require.ErrorIs(t, err, myerrors.ErrNotFound)
	})

	t.Run("repository error", func(t *testing.T) {
		currentUser := schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)

		cr.EXPECT().Get(tx, contestID).Return(nil, errors.New("db error")).Times(1)

		result, err := cs.Get(tx, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "db error", err.Error())
	})
}
