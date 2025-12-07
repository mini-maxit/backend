package service_test

import (
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	repository "github.com/mini-maxit/backend/package/repository"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestContestService_GetMyContestResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
	db := &testutils.MockDatabase{}

	visibleTrue := true
	visibleFalse := false

	t.Run("contest not found", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleStudent}
		contestID := int64(999)

		cr.EXPECT().Get(db, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		result, err := cs.GetMyContestResults(db, currentUser, contestID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNotFound)
		assert.Nil(t, result)
	})

	t.Run("user without access (invisible + not participant + no permission)", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleStudent}
		contestID := int64(55)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Hidden Contest",
			CreatedBy: 2,
			IsVisible: visibleFalse,
		}

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(false, nil).Times(1)

		result, err := cs.GetMyContestResults(db, currentUser, contestID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
		assert.Nil(t, result)
	})

	t.Run("empty task list", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleStudent}
		contestID := int64(100)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Empty Tasks Contest",
			CreatedBy: 5,
			IsVisible: visibleTrue,
		}

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		cr.EXPECT().GetTasksForContest(db, contestID).Return([]models.Task{}, nil).Times(1)
		sr.EXPECT().GetAllTaskStatsForContestUser(db, contestID, currentUser.ID).Return([]repository.TaskUserSubmissionStats{}, nil).Times(1)

		result, err := cs.GetMyContestResults(db, currentUser, contestID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, contestID, result.Contest.ID)
		assert.Empty(t, result.TaskResults)
	})

	t.Run("tasks with no submissions (zero attempts)", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleStudent}
		contestID := int64(200)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "No Submissions Contest",
			CreatedBy: 7,
			IsVisible: visibleTrue,
		}

		task := models.Task{ID: 1, Title: "Task A"}
		tasks := []models.Task{task}

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		cr.EXPECT().GetTasksForContest(db, contestID).Return(tasks, nil).Times(1)
		stats := []repository.TaskUserSubmissionStats{
			{TaskID: task.ID, AttemptCount: 0, BestScore: 0, BestSubmissionID: nil},
		}
		sr.EXPECT().GetAllTaskStatsForContestUser(db, contestID, currentUser.ID).Return(stats, nil).Times(1)

		result, err := cs.GetMyContestResults(db, currentUser, contestID)
		require.NoError(t, err)
		require.Len(t, result.TaskResults, 1)
		assert.Equal(t, 0, result.TaskResults[0].SubmissionCount)
		assert.Equal(t, 0.0, result.TaskResults[0].BestScore) //nolint:testifylint // direct comparison with float64
		assert.Nil(t, result.TaskResults[0].BestSubmissionID)
	})

	t.Run("tasks with submissions and best score calculation", func(t *testing.T) {
		currentUser := &schemas.User{ID: 15, Role: types.UserRoleStudent}
		contestID := int64(300)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Submissions Contest",
			CreatedBy: 9,
			IsVisible: visibleTrue,
		}

		task1 := models.Task{ID: 11, Title: "Task One"}
		task2 := models.Task{ID: 12, Title: "Task Two"}
		tasks := []models.Task{task1, task2}

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		cr.EXPECT().GetTasksForContest(db, contestID).Return(tasks, nil).Times(1)

		bestSub1 := int64(1001)
		bestSub2 := int64(1002)
		stats := []repository.TaskUserSubmissionStats{
			{TaskID: task1.ID, AttemptCount: 5, BestScore: 80.0, BestSubmissionID: &bestSub1},
			{TaskID: task2.ID, AttemptCount: 3, BestScore: 66.6667, BestSubmissionID: &bestSub2},
		}
		sr.EXPECT().GetAllTaskStatsForContestUser(db, contestID, currentUser.ID).Return(stats, nil).Times(1)

		result, err := cs.GetMyContestResults(db, currentUser, contestID)
		require.NoError(t, err)
		require.Len(t, result.TaskResults, 2)

		tr1 := result.TaskResults[0]
		tr2 := result.TaskResults[1]

		assert.Equal(t, int64(11), tr1.Task.ID)
		assert.Equal(t, 5, tr1.SubmissionCount)
		assert.InDelta(t, 80.0, tr1.BestScore, 0.001)
		require.NotNil(t, tr1.BestSubmissionID)
		assert.Equal(t, bestSub1, *tr1.BestSubmissionID)

		assert.Equal(t, int64(12), tr2.Task.ID)
		assert.Equal(t, 3, tr2.SubmissionCount)
		assert.InDelta(t, 66.6667, tr2.BestScore, 0.001)
		require.NotNil(t, tr2.BestSubmissionID)
		assert.Equal(t, bestSub2, *tr2.BestSubmissionID)
	})

	t.Run("creator not preloaded - still succeeds", func(t *testing.T) {
		currentUser := &schemas.User{ID: 22, Role: types.UserRoleStudent}
		contestID := int64(400)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "No Creator Preload",
			CreatedBy: 22,
			IsVisible: visibleTrue,
		}

		task := models.Task{ID: 50, Title: "Solo Task"}
		tasks := []models.Task{task}

		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		cr.EXPECT().GetTasksForContest(db, contestID).Return(tasks, nil).Times(1)
		stats := []repository.TaskUserSubmissionStats{
			{TaskID: task.ID, AttemptCount: 1, BestScore: 100.0, BestSubmissionID: nil},
		}
		sr.EXPECT().GetAllTaskStatsForContestUser(db, contestID, currentUser.ID).Return(stats, nil).Times(1)

		result, err := cs.GetMyContestResults(db, currentUser, contestID)
		require.NoError(t, err)
		require.Len(t, result.TaskResults, 1)
		assert.Equal(t, contest.CreatedBy, result.Contest.CreatedBy)
		assert.Equal(t, "No Creator Preload", result.Contest.Name)
	})
}

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
			IsVisible:   visible,
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
			IsVisible:   visible,
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
				IsVisible:   visible,
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
				IsVisible:   visible,
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

func TestContestService_GetPastContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
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
					IsVisible: visible,
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
	gr := mock_repository.NewMockGroupRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
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
					IsVisible: visible,
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
	gr := mock_repository.NewMockGroupRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
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
	gr := mock_repository.NewMockGroupRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
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

func TestContestService_GetDetailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("successful retrieval - visible contest", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		visible := true

		contest := &repository.ContestDetailed{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: visible,
			},
		}

		cr.EXPECT().GetDetailed(db, contestID).Return(contest, nil).Times(1)

		result, err := cs.GetDetailed(db, currentUser, contestID)

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

		contest := &repository.ContestDetailed{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: visible,
			},
		}

		cr.EXPECT().Get(db, contestID).Return(&contest.Contest, nil).Times(1)
		cr.EXPECT().GetDetailed(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		result, err := cs.GetDetailed(db, currentUser, contestID)

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

		contest := &repository.ContestDetailed{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: visible,
			},
		}

		cr.EXPECT().Get(db, contestID).Return(&contest.Contest, nil).Times(1)
		cr.EXPECT().GetDetailed(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)

		result, err := cs.GetDetailed(db, currentUser, contestID)

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

		contest := &repository.ContestDetailed{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: visible,
			},
		}

		cr.EXPECT().Get(db, contestID).Return(&contest.Contest, nil).Times(1)
		cr.EXPECT().GetDetailed(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(false, nil).Times(1)

		result, err := cs.GetDetailed(db, currentUser, contestID)

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

		contest := &repository.ContestDetailed{
			Contest: models.Contest{
				ID:        contestID,
				Name:      "Test Contest",
				CreatedBy: 2,
				IsVisible: visible,
			},
		}

		cr.EXPECT().Get(db, contestID).Return(&contest.Contest, nil).Times(1)
		cr.EXPECT().GetDetailed(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(false, nil).Times(1)

		result, err := cs.GetDetailed(db, currentUser, contestID)

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

		cr.EXPECT().GetDetailed(db, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		result, err := cs.GetDetailed(db, currentUser, contestID)

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

		cr.EXPECT().GetDetailed(db, contestID).Return(nil, errors.ErrDatabaseConnection).Times(1)

		result, err := cs.GetDetailed(db, currentUser, contestID)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, errors.ErrDatabaseConnection.Message, err.Error())
	})
}

func TestContestService_GetVisibleTasksForContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
	db := &testutils.MockDatabase{}

	visible := true

	t.Run("successful retrieval - participant", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: visible,
		}

		// hasContestPermission calls Get first
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(true, nil).Times(1)

		tasks := []models.ContestTask{
			{
				ContestID: contestID,
				TaskID:    1,
				Task: models.Task{
					ID:    1,
					Title: "Test Task",
					Author: models.User{
						Name: "Creator",
					},
				},
				StartAt:          time.Now().Add(-1 * time.Hour),
				EndAt:            nil,
				IsSubmissionOpen: true,
			},
		}

		cr.EXPECT().GetVisibleContestTasksWithSettings(db, contestID).Return(tasks, nil).Times(1)

		result, err := cs.GetVisibleTasksForContest(db, currentUser, contestID, types.ContestStatusOngoing)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), result[0].Task.ID)
		assert.Equal(t, "Test Task", result[0].Task.Title)
		assert.Equal(t, "Creator", result[0].CreatorName)
	})

	t.Run("successful retrieval - user with edit permission", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleTeacher,
		}
		contestID := int64(10)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: visible,
		}

		// hasContestPermission calls Get first
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(false, nil).Times(1)

		tasks := []models.ContestTask{
			{
				ContestID: contestID,
				TaskID:    1,
				Task: models.Task{
					ID:    1,
					Title: "Test Task",
					Author: models.User{
						Name: "Creator",
					},
				},
				StartAt:          time.Now().Add(-1 * time.Hour),
				EndAt:            nil,
				IsSubmissionOpen: true,
			},
		}

		cr.EXPECT().GetVisibleContestTasksWithSettings(db, contestID).Return(tasks, nil).Times(1)

		result, err := cs.GetVisibleTasksForContest(db, currentUser, contestID, types.ContestStatusOngoing)

		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("unauthorized - not participant and no permission", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: visible,
		}

		// hasContestPermission calls Get first
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(false, nil).Times(1)

		result, err := cs.GetVisibleTasksForContest(db, currentUser, contestID, types.ContestStatusOngoing)

		require.Error(t, err)
		assert.Nil(t, result)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("filter by status - ongoing", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		now := time.Now()

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: visible,
		}

		// hasContestPermission calls Get first
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(true, nil).Times(1)

		pastEnd := now.Add(-1 * time.Hour)
		tasks := []models.ContestTask{
			{
				ContestID: contestID,
				TaskID:    1,
				Task: models.Task{
					ID:    1,
					Title: "Ongoing Task",
					Author: models.User{
						Name: "Creator",
					},
				},
				StartAt:          now.Add(-2 * time.Hour),
				EndAt:            nil, // No end = ongoing
				IsSubmissionOpen: true,
			},
			{
				ContestID: contestID,
				TaskID:    2,
				Task: models.Task{
					ID:    2,
					Title: "Past Task",
					Author: models.User{
						Name: "Creator",
					},
				},
				StartAt:          now.Add(-3 * time.Hour),
				EndAt:            &pastEnd, // Ended
				IsSubmissionOpen: false,
			},
		}

		cr.EXPECT().GetVisibleContestTasksWithSettings(db, contestID).Return(tasks, nil).Times(1)

		result, err := cs.GetVisibleTasksForContest(db, currentUser, contestID, types.ContestStatusOngoing)

		require.NoError(t, err)
		assert.Len(t, result, 1) // Only ongoing task
		assert.Equal(t, "Ongoing Task", result[0].Task.Title)
	})

	t.Run("filter by status - past", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		now := time.Now()

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: visible,
		}

		// hasContestPermission calls Get first
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(true, nil).Times(1)

		pastEnd := now.Add(-1 * time.Hour)
		tasks := []models.ContestTask{
			{
				ContestID: contestID,
				TaskID:    1,
				Task: models.Task{
					ID:    1,
					Title: "Ongoing Task",
					Author: models.User{
						Name: "Creator",
					},
				},
				StartAt:          now.Add(-2 * time.Hour),
				EndAt:            nil,
				IsSubmissionOpen: true,
			},
			{
				ContestID: contestID,
				TaskID:    2,
				Task: models.Task{
					ID:    2,
					Title: "Past Task",
					Author: models.User{
						Name: "Creator",
					},
				},
				StartAt:          now.Add(-3 * time.Hour),
				EndAt:            &pastEnd,
				IsSubmissionOpen: false,
			},
		}

		cr.EXPECT().GetVisibleContestTasksWithSettings(db, contestID).Return(tasks, nil).Times(1)

		result, err := cs.GetVisibleTasksForContest(db, currentUser, contestID, types.ContestStatusPast)

		require.NoError(t, err)
		assert.Len(t, result, 1) // Only past task
		assert.Equal(t, "Past Task", result[0].Task.Title)
	})

	t.Run("filter by status - upcoming", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)
		now := time.Now()

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: visible,
		}

		// hasContestPermission calls Get first
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(true, nil).Times(1)

		futureStart := now.Add(1 * time.Hour)
		tasks := []models.ContestTask{
			{
				ContestID: contestID,
				TaskID:    1,
				Task: models.Task{
					ID:    1,
					Title: "Upcoming Task",
					Author: models.User{
						Name: "Creator",
					},
				},
				StartAt:          futureStart,
				EndAt:            nil,
				IsSubmissionOpen: false,
			},
			{
				ContestID: contestID,
				TaskID:    2,
				Task: models.Task{
					ID:    2,
					Title: "Ongoing Task",
					Author: models.User{
						Name: "Creator",
					},
				},
				StartAt:          now.Add(-1 * time.Hour),
				EndAt:            nil,
				IsSubmissionOpen: true,
			},
		}

		cr.EXPECT().GetVisibleContestTasksWithSettings(db, contestID).Return(tasks, nil).Times(1)

		result, err := cs.GetVisibleTasksForContest(db, currentUser, contestID, types.ContestStatusUpcoming)

		require.NoError(t, err)
		assert.Len(t, result, 1) // Only upcoming task
		assert.Equal(t, "Upcoming Task", result[0].Task.Title)
	})

	t.Run("repository error", func(t *testing.T) {
		currentUser := &schemas.User{
			ID:   1,
			Role: types.UserRoleStudent,
		}
		contestID := int64(10)

		contest := &models.Contest{
			ID:        contestID,
			Name:      "Test Contest",
			CreatedBy: 2,
			IsVisible: visible,
		}

		// hasContestPermission calls Get first
		cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)
		cr.EXPECT().IsUserParticipant(db, contestID, currentUser.ID).Return(true, nil).Times(1)
		cr.EXPECT().GetVisibleContestTasksWithSettings(db, contestID).Return(nil, errors.ErrDatabaseConnection).Times(1)

		result, err := cs.GetVisibleTasksForContest(db, currentUser, contestID, types.ContestStatusOngoing)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestContestService_AddGroupToContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("no permission", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleStudent}
		contestID := int64(1)
		groupID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

		err := cs.AddGroupToContest(db, currentUser, contestID, groupID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("contest not found", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleAdmin}
		contestID := int64(999)
		groupID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.AddGroupToContest(db, currentUser, contestID, groupID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("group not found", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleAdmin}
		contestID := int64(1)
		groupID := int64(999)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		gr.EXPECT().Get(db, groupID).Return(nil, gorm.ErrRecordNotFound).Times(1)

		err := cs.AddGroupToContest(db, currentUser, contestID, groupID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleTeacher}
		contestID := int64(1)
		groupID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, CreatedBy: 10}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		gr.EXPECT().Get(db, groupID).Return(&models.Group{ID: groupID}, nil).Times(1)
		cr.EXPECT().GetContestGroups(db, contestID).Return([]models.Group{}, nil).Times(1)
		cr.EXPECT().AddGroupToContest(db, contestID, groupID).Return(nil).Times(1)

		err := cs.AddGroupToContest(db, currentUser, contestID, groupID)
		require.NoError(t, err)
	})

	t.Run("group already assigned", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleTeacher}
		contestID := int64(1)
		groupID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, CreatedBy: 10}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		gr.EXPECT().Get(db, groupID).Return(&models.Group{ID: groupID}, nil).Times(1)
		cr.EXPECT().GetContestGroups(db, contestID).Return([]models.Group{{ID: groupID}}, nil).Times(1)

		err := cs.AddGroupToContest(db, currentUser, contestID, groupID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrGroupAlreadyAssignedToContest)
	})
}

func TestContestService_RemoveGroupFromContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("no permission", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleStudent}
		contestID := int64(1)
		groupID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

		err := cs.RemoveGroupFromContest(db, currentUser, contestID, groupID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("success", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleTeacher}
		contestID := int64(1)
		groupID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, CreatedBy: 10}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().GetContestGroups(db, contestID).Return([]models.Group{{ID: groupID}}, nil).Times(1)
		cr.EXPECT().RemoveGroupFromContest(db, contestID, groupID).Return(nil).Times(1)

		err := cs.RemoveGroupFromContest(db, currentUser, contestID, groupID)
		require.NoError(t, err)
	})

	t.Run("group not assigned", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleTeacher}
		contestID := int64(1)
		groupID := int64(5)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, CreatedBy: 10}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().GetContestGroups(db, contestID).Return([]models.Group{}, nil).Times(1)

		err := cs.RemoveGroupFromContest(db, currentUser, contestID, groupID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrGroupNotAssignedToContest)
	})
}

func TestContestService_GetContestGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("no permission", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleStudent}
		contestID := int64(1)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

		result, err := cs.GetContestGroups(db, currentUser, contestID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
		assert.Nil(t, result)
	})

	t.Run("success with groups", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleTeacher}
		contestID := int64(1)

		assignedGroups := []models.Group{
			{ID: 1, Name: "Group 1", CreatedBy: 10},
			{ID: 2, Name: "Group 2", CreatedBy: 10},
		}

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, CreatedBy: 10}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().GetContestGroups(db, contestID).Return(assignedGroups, nil).Times(1)

		result, err := cs.GetContestGroups(db, currentUser, contestID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Group 1", result[0].Name)
		assert.Equal(t, int64(2), result[1].ID)
		assert.Equal(t, "Group 2", result[1].Name)
	})

	t.Run("success with empty groups", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleAdmin}
		contestID := int64(1)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().GetContestGroups(db, contestID).Return([]models.Group{}, nil).Times(1)

		result, err := cs.GetContestGroups(db, currentUser, contestID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Empty(t, result)
	})
}

func TestContestService_GetAssignableGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
	db := &testutils.MockDatabase{}

	t.Run("no permission", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleStudent}
		contestID := int64(1)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

		result, err := cs.GetAssignableGroups(db, currentUser, contestID)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrForbidden)
		assert.Nil(t, result)
	})

	t.Run("success with groups", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleTeacher}
		contestID := int64(1)

		assignableGroups := []models.Group{
			{ID: 3, Name: "Group 3", CreatedBy: 10},
			{ID: 4, Name: "Group 4", CreatedBy: 10},
		}

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, CreatedBy: 10}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().GetAssignableGroups(db, contestID).Return(assignableGroups, nil).Times(1)

		result, err := cs.GetAssignableGroups(db, currentUser, contestID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(3), result[0].ID)
		assert.Equal(t, "Group 3", result[0].Name)
		assert.Equal(t, int64(4), result[1].ID)
		assert.Equal(t, "Group 4", result[1].Name)
	})

	t.Run("success with empty groups", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleAdmin}
		contestID := int64(1)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().GetAssignableGroups(db, contestID).Return([]models.Group{}, nil).Times(1)

		result, err := cs.GetAssignableGroups(db, currentUser, contestID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Empty(t, result)
	})
}
