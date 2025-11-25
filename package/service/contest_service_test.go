package service_test

import (
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
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
			StartAt:     startTime,
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
	assert.Equal(t, startTime, result.StartAt)
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
			StartAt:     startTime,
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
	assert.Equal(t, startTime, result.StartAt)
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
				StartAt:     startTime,
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
				StartAt:     startTime,
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
			StartAt:     startTime,
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
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("successful retrieval", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().GetPastContestsWithStats(db, currentUser.ID, 0, 10, "start_time").Return(contestsWithStats, int64(1), nil).Times(1)

		result, err := cs.GetPastContests(db, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, int64(1), result.Items[0].ID)
		assert.Equal(t, 1, result.Pagination.TotalItems)
		assert.Equal(t, 1, result.Pagination.CurrentPage)
		assert.Equal(t, 10, result.Pagination.PageSize)
	})

	t.Run("repository error", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		queryParams := schemas.PaginationParams{
			Limit:  10,
			Offset: 0,
			Sort:   "start_time",
		}

		cr.EXPECT().GetPastContestsWithStats(db, currentUser.ID, 0, 10, "start_time").Return(nil, int64(0), errors.ErrDatabaseConnection).Times(1)

		result, err := cs.GetPastContests(db, currentUser, queryParams)

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
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("successful retrieval", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().GetUpcomingContestsWithStats(db, currentUser.ID, 0, 10, "start_time").Return(contestsWithStats, int64(1), nil).Times(1)

		result, err := cs.GetUpcomingContests(db, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, int64(1), result.Items[0].ID)
		assert.Equal(t, 1, result.Pagination.TotalItems)
		assert.Equal(t, 1, result.Pagination.CurrentPage)
		assert.Equal(t, 10, result.Pagination.PageSize)
	})

	t.Run("repository error", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		queryParams := schemas.PaginationParams{
			Limit:  10,
			Offset: 0,
			Sort:   "start_time",
		}

		cr.EXPECT().GetUpcomingContestsWithStats(db, currentUser.ID, 0, 10, "start_time").Return(nil, int64(0), errors.ErrDatabaseConnection).Times(1)

		result, err := cs.GetUpcomingContests(db, currentUser, queryParams)

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
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("successful approval by admin", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusApproved).Return(nil).Times(1)
		cr.EXPECT().CreateContestParticipant(db, contestID, userID).Return(nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.NoError(t, err)
	})

	t.Run("successful approval by contest creator teacher", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusApproved).Return(nil).Times(1)
		cr.EXPECT().CreateContestParticipant(db, contestID, userID).Return(nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.NoError(t, err)
	})

	t.Run("unauthorized - student role", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		userID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{
			ID: contestID,
		}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("contest not found", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(999)
		userID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("teacher not authorized - not contest creator", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   3,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		userID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{
			ID: contestID,
		}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("user not found", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("user already participant", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(true, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().DeleteRegistrationRequest(db, request.ID).Return(nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrAlreadyParticipant)
	})

	t.Run("no pending registration", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(nil, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNoPendingRegistration)
	})

	t.Run("error checking participant status", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, gorm.ErrInvalidDB).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})

	t.Run("error checking pending registration", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(nil, gorm.ErrInvalidDB).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})

	t.Run("error updating registration status", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusApproved).Return(gorm.ErrInvalidDB).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})

	t.Run("error adding user as participant", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusApproved).Return(nil).Times(1)
		cr.EXPECT().CreateContestParticipant(db, contestID, userID).Return(gorm.ErrInvalidDB).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.ApproveRegistrationRequest(db, currentUser, contestID, userID)

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
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("successful rejection by admin", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusRejected).Return(nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

		require.NoError(t, err)
	})

	t.Run("successful rejection by contest creator teacher", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusRejected).Return(nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

		require.NoError(t, err)
	})

	t.Run("unauthorized - student role", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		userID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{
			ID: contestID,
		}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("contest not found", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(999)
		userID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("teacher not authorized", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   3,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		userID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{
			ID: contestID,
		}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("user not found", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("user already participant", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(true, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().DeleteRegistrationRequest(db, request.ID).Return(nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrAlreadyParticipant)
	})

	t.Run("no pending registration", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(nil, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNoPendingRegistration)
	})

	t.Run("error updating registration status", func(t *testing.T) {
		currentUser := &schemas.User{
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

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		ur.EXPECT().Get(db, userID).Return(user, nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil).Times(1)
		cr.EXPECT().GetPendingRegistrationRequest(db, contestID, userID).Return(request, nil).Times(1)
		cr.EXPECT().UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusRejected).Return(gorm.ErrInvalidDB).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		err := cs.RejectRegistrationRequest(db, currentUser, contestID, userID)

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
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("successful retrieval - visible contest", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		visible := true

		contest := &models.ParticipantContestStats{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: &visible,
			},
		}

		cr.EXPECT().GetWithCount(db, contestID).Return(contest, nil).Times(1)

		result, err := cs.Get(db, currentUser, contestID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, contestID, result.ID)
		assert.Equal(t, "Test Contest", result.Name)
	})

	t.Run("successful retrieval - invisible contest by admin", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleAdmin,
		}
		contestID := int64(10)
		visible := false

		contest := &models.ParticipantContestStats{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: &visible,
			},
		}

		cr.EXPECT().GetWithCount(db, contestID).Return(contest, nil).Times(1)

		result, err := cs.Get(db, currentUser, contestID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, contestID, result.ID)
	})

	t.Run("successful retrieval - invisible contest by creator teacher", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   2,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		visible := false

		contest := &models.ParticipantContestStats{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: &visible,
			},
		}

		cr.EXPECT().Get(db, contestID).Return(&contest.Contest, nil).Times(1)
		cr.EXPECT().GetWithCount(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		result, err := cs.Get(db, currentUser, contestID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, contestID, result.ID)
	})

	t.Run("unauthorized - student viewing invisible contest", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		visible := false

		contest := &models.ParticipantContestStats{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: &visible,
			},
		}

		cr.EXPECT().Get(db, contestID).Return(&contest.Contest, nil).Times(1)
		cr.EXPECT().GetWithCount(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(false, nil).Times(1)

		result, err := cs.Get(db, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("unauthorized - teacher viewing invisible contest not created by them", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   3,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)
		visible := false

		contest := &models.ParticipantContestStats{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: &visible,
			},
		}

		cr.EXPECT().Get(db, contestID).Return(&contest.Contest, nil).Times(1)
		cr.EXPECT().GetWithCount(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(false, nil).Times(1)

		result, err := cs.Get(db, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("contest not found", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(999)

		cr.EXPECT().GetWithCount(db, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		result, err := cs.Get(db, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("repository error", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)

		cr.EXPECT().GetWithCount(db, contestID).Return(nil, errors.ErrDatabaseConnection).Times(1)

		result, err := cs.Get(db, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, errors.ErrDatabaseConnection.Message, err.Error())
	})
}
