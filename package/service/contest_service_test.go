package service_test

import (
	"testing"
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestContestWithStatsToSchema(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	pastTime := now.Add(-24 * time.Hour)

	t.Run("converts contest with stats to schema - can register", func(t *testing.T) {
		contestWithStats := &models.ContestWithStats{
			Contest: models.Contest{
				ID:                 1,
				Name:               "Test Contest",
				Description:        "Test Description",
				CreatedBy:          1,
				StartAt:            &now,
				EndAt:              &futureTime,
				IsRegistrationOpen: func() *bool { b := true; return &b }(),
				IsSubmissionOpen:   func() *bool { b := false; return &b }(),
				IsVisible:          func() *bool { b := true; return &b }(),
				BaseModel: models.BaseModel{
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			ParticipantCount: 5,
			IsParticipant:    false,
			HasPendingReg:    false,
		}

		result := service.ContestWithStatsToSchema(contestWithStats)

		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, "Test Contest", result.Name)
		assert.Equal(t, "Test Description", result.Description)
		assert.Equal(t, int64(1), result.CreatedBy)
		assert.Equal(t, &now, result.StartAt)
		assert.Equal(t, &futureTime, result.EndAt)
		assert.Equal(t, int64(5), result.ParticipantCount)
		assert.Equal(t, "canRegister", result.RegistrationStatus)
	})

	t.Run("converts contest with stats to schema - registered", func(t *testing.T) {
		contestWithStats := &models.ContestWithStats{
			Contest: models.Contest{
				ID:                 2,
				Name:               "Test Contest 2",
				Description:        "Test Description 2",
				CreatedBy:          2,
				EndAt:              &futureTime,
				IsRegistrationOpen: func() *bool { b := true; return &b }(),
				IsSubmissionOpen:   func() *bool { b := true; return &b }(),
				IsVisible:          func() *bool { b := true; return &b }(),
			},
			ParticipantCount: 10,
			TaskCount:        7,
			IsParticipant:    true,
			HasPendingReg:    false,
		}

		result := service.ContestWithStatsToSchema(contestWithStats)

		assert.Equal(t, int64(10), result.ParticipantCount)
		assert.Equal(t, int64(7), result.TaskCount)
		assert.Equal(t, "registered", result.RegistrationStatus)
	})

	t.Run("converts contest with stats to schema - awaiting approval", func(t *testing.T) {
		contestWithStats := &models.ContestWithStats{
			Contest: models.Contest{
				ID:                 3,
				Name:               "Test Contest 3",
				Description:        "Test Description 3",
				CreatedBy:          3,
				EndAt:              &futureTime,
				IsRegistrationOpen: func() *bool { b := true; return &b }(),
				IsSubmissionOpen:   func() *bool { b := false; return &b }(),
				IsVisible:          func() *bool { b := true; return &b }(),
			},
			ParticipantCount: 3,
			TaskCount:        2,
			IsParticipant:    false,
			HasPendingReg:    true,
		}

		result := service.ContestWithStatsToSchema(contestWithStats)

		assert.Equal(t, int64(3), result.ParticipantCount)
		assert.Equal(t, int64(2), result.TaskCount)
		assert.Equal(t, "awaitingApproval", result.RegistrationStatus)
	})

	t.Run("converts contest with stats to schema - registration closed (registration disabled)", func(t *testing.T) {
		contestWithStats := &models.ContestWithStats{
			Contest: models.Contest{
				ID:                 4,
				Name:               "Test Contest 4",
				Description:        "Test Description 4",
				CreatedBy:          4,
				EndAt:              &futureTime,
				IsRegistrationOpen: func() *bool { b := false; return &b }(),
				IsSubmissionOpen:   func() *bool { b := false; return &b }(),
				IsVisible:          func() *bool { b := true; return &b }(),
			},
			ParticipantCount: 3,
			TaskCount:        1,
			IsParticipant:    false,
			HasPendingReg:    false,
		}

		result := service.ContestWithStatsToSchema(contestWithStats)

		assert.Equal(t, int64(3), result.ParticipantCount)
		assert.Equal(t, int64(1), result.TaskCount)
		assert.Equal(t, "registrationClosed", result.RegistrationStatus)
	})

	t.Run("converts contest with stats to schema - registration closed (contest ended)", func(t *testing.T) {
		contestWithStats := &models.ContestWithStats{
			Contest: models.Contest{
				ID:                 5,
				Name:               "Test Contest 5",
				Description:        "Test Description 5",
				CreatedBy:          5,
				EndAt:              &pastTime, // Contest has ended
				IsRegistrationOpen: func() *bool { b := true; return &b }(),
				IsSubmissionOpen:   func() *bool { b := false; return &b }(),
				IsVisible:          func() *bool { b := true; return &b }(),
			},
			ParticipantCount: 3,
			TaskCount:        4,
			IsParticipant:    false,
			HasPendingReg:    false,
		}

		result := service.ContestWithStatsToSchema(contestWithStats)

		assert.Equal(t, int64(3), result.ParticipantCount)
		assert.Equal(t, int64(4), result.TaskCount)
		assert.Equal(t, "registrationClosed", result.RegistrationStatus)
	})

	t.Run("converts contest with stats to schema - registered even if contest ended", func(t *testing.T) {
		contestWithStats := &models.ContestWithStats{
			Contest: models.Contest{
				ID:                 6,
				Name:               "Test Contest 6",
				Description:        "Test Description 6",
				CreatedBy:          6,
				EndAt:              &pastTime, // Contest has ended
				IsRegistrationOpen: func() *bool { b := true; return &b }(),
				IsSubmissionOpen:   func() *bool { b := false; return &b }(),
				IsVisible:          func() *bool { b := true; return &b }(),
			},
			ParticipantCount: 3,
			IsParticipant:    true, // User is already registered
			HasPendingReg:    false,
		}

		result := service.ContestWithStatsToSchema(contestWithStats)

		assert.Equal(t, int64(3), result.ParticipantCount)
		assert.Equal(t, "registered", result.RegistrationStatus)
	})

	t.Run("converts contest with stats to schema - no end date, registration open", func(t *testing.T) {
		contestWithStats := &models.ContestWithStats{
			Contest: models.Contest{
				ID:                 7,
				Name:               "Test Contest 7",
				Description:        "Test Description 7",
				CreatedBy:          7,
				EndAt:              nil, // No end date
				IsRegistrationOpen: func() *bool { b := true; return &b }(),
				IsSubmissionOpen:   func() *bool { b := false; return &b }(),
				IsVisible:          func() *bool { b := true; return &b }(),
			},
			ParticipantCount: 3,
			TaskCount:        6,
			IsParticipant:    false,
			HasPendingReg:    false,
		}

		result := service.ContestWithStatsToSchema(contestWithStats)

		assert.Equal(t, int64(3), result.ParticipantCount)
		assert.Equal(t, int64(6), result.TaskCount)
		assert.Equal(t, "canRegister", result.RegistrationStatus)
	})
}

func TestContestService_GetAll_WithStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr)
	tx := &gorm.DB{}

	currentUser := schemas.User{
		ID:   1,
		Role: types.UserRoleStudent,
	}

	queryParams := map[string]any{
		"limit":  10,
		"offset": 0,
		"sort":   "created_at:desc",
	}

	t.Run("successful GetAll with stats", func(t *testing.T) {
		now := time.Now()
		futureTime := now.Add(24 * time.Hour)
		contestsWithStats := []models.ContestWithStats{
			{
				Contest: models.Contest{
					ID:                 1,
					Name:               "Contest 1",
					Description:        "Description 1",
					CreatedBy:          2,
					EndAt:              &futureTime,
					IsRegistrationOpen: func() *bool { b := true; return &b }(),
					IsSubmissionOpen:   func() *bool { b := false; return &b }(),
					IsVisible:          func() *bool { b := true; return &b }(),
					BaseModel: models.BaseModel{
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				ParticipantCount: 5,
				TaskCount:        3,
				IsParticipant:    false,
				HasPendingReg:    true,
			},
			{
				Contest: models.Contest{
					ID:                 2,
					Name:               "Contest 2",
					Description:        "Description 2",
					CreatedBy:          2,
					EndAt:              &futureTime,
					IsRegistrationOpen: func() *bool { b := true; return &b }(),
					IsSubmissionOpen:   func() *bool { b := true; return &b }(),
					IsVisible:          func() *bool { b := true; return &b }(),
					BaseModel: models.BaseModel{
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				ParticipantCount: 10,
				TaskCount:        7,
				IsParticipant:    true,
				HasPendingReg:    false,
			},
		}

		cr.EXPECT().GetAllWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(contestsWithStats, nil).Times(1)

		result, err := cs.GetAll(tx, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Check first contest
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Contest 1", result[0].Name)
		assert.Equal(t, int64(5), result[0].ParticipantCount)
		assert.Equal(t, int64(3), result[0].TaskCount)
		assert.Equal(t, "awaitingApproval", result[0].RegistrationStatus)

		// Check second contest
		assert.Equal(t, int64(2), result[1].ID)
		assert.Equal(t, "Contest 2", result[1].Name)
		assert.Equal(t, int64(10), result[1].ParticipantCount)
		assert.Equal(t, int64(7), result[1].TaskCount)
		assert.Equal(t, "registered", result[1].RegistrationStatus)
	})

	t.Run("repository error", func(t *testing.T) {
		cr.EXPECT().GetAllWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(nil, gorm.ErrInvalidDB).Times(1)

		result, err := cs.GetAll(tx, currentUser, queryParams)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})

	t.Run("filters invisible contests for students", func(t *testing.T) {
		now := time.Now()
		futureTime := now.Add(24 * time.Hour)
		contestsWithStats := []models.ContestWithStats{
			{
				Contest: models.Contest{
					ID:                 1,
					Name:               "Visible Contest",
					Description:        "Description 1",
					CreatedBy:          2,
					EndAt:              &futureTime,
					IsRegistrationOpen: func() *bool { b := true; return &b }(),
					IsSubmissionOpen:   func() *bool { b := false; return &b }(),
					IsVisible:          func() *bool { b := true; return &b }(),
					BaseModel: models.BaseModel{
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				ParticipantCount: 5,
				TaskCount:        2,
				IsParticipant:    false,
				HasPendingReg:    false,
			},
			{
				Contest: models.Contest{
					ID:                 2,
					Name:               "Invisible Contest",
					Description:        "Description 2",
					CreatedBy:          2,
					EndAt:              &futureTime,
					IsRegistrationOpen: func() *bool { b := true; return &b }(),
					IsSubmissionOpen:   func() *bool { b := true; return &b }(),
					IsVisible:          func() *bool { b := false; return &b }(),
					BaseModel: models.BaseModel{
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				ParticipantCount: 3,
				TaskCount:        1,
				IsParticipant:    false,
				HasPendingReg:    false,
			},
		}

		cr.EXPECT().GetAllWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(contestsWithStats, nil).Times(1)
		cr.EXPECT().IsUserParticipant(tx, int64(2), currentUser.ID).Return(false, nil).Times(1)

		result, err := cs.GetAll(tx, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Visible Contest", result[0].Name)
		assert.Equal(t, "canRegister", result[0].RegistrationStatus)
	})
}

func TestContestService_GetOngoingContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr)
	tx := &gorm.DB{}

	currentUser := schemas.User{
		ID:   1,
		Role: types.UserRoleStudent,
	}

	queryParams := map[string]any{
		"limit":  10,
		"offset": 0,
		"sort":   "created_at:desc",
	}

	t.Run("successful GetOngoingContests", func(t *testing.T) {
		now := time.Now()
		futureTime := now.Add(24 * time.Hour)
		ongoingContestsWithStats := []models.ContestWithStats{
			{
				Contest: models.Contest{
					ID:                 1,
					Name:               "Ongoing Contest 1",
					Description:        "Description 1",
					CreatedBy:          2,
					StartAt:            &now,
					EndAt:              &futureTime,
					IsRegistrationOpen: func() *bool { b := true; return &b }(),
					IsSubmissionOpen:   func() *bool { b := true; return &b }(),
					IsVisible:          func() *bool { b := true; return &b }(),
					BaseModel: models.BaseModel{
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				ParticipantCount: 8,
				TaskCount:        4,
				IsParticipant:    true,
				HasPendingReg:    false,
			},
		}

		cr.EXPECT().GetOngoingContestsWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(ongoingContestsWithStats, nil).Times(1)

		result, err := cs.GetOngoingContests(tx, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Ongoing Contest 1", result[0].Name)
		assert.Equal(t, int64(8), result[0].ParticipantCount)
		assert.Equal(t, int64(4), result[0].TaskCount)
		assert.Equal(t, "registered", result[0].RegistrationStatus)
	})

	t.Run("repository error", func(t *testing.T) {
		cr.EXPECT().GetOngoingContestsWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(nil, gorm.ErrInvalidDB).Times(1)

		result, err := cs.GetOngoingContests(tx, currentUser, queryParams)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})
}

func TestContestService_GetPastContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr)
	tx := &gorm.DB{}

	currentUser := schemas.User{
		ID:   1,
		Role: types.UserRoleStudent,
	}

	queryParams := map[string]any{
		"limit":  10,
		"offset": 0,
		"sort":   "created_at:desc",
	}

	t.Run("successful GetPastContests", func(t *testing.T) {
		now := time.Now()
		pastTime := now.Add(-24 * time.Hour)
		pastContestsWithStats := []models.ContestWithStats{
			{
				Contest: models.Contest{
					ID:                 1,
					Name:               "Past Contest 1",
					Description:        "Description 1",
					CreatedBy:          2,
					StartAt:            &pastTime,
					EndAt:              &pastTime,
					IsRegistrationOpen: func() *bool { b := false; return &b }(),
					IsSubmissionOpen:   func() *bool { b := false; return &b }(),
					IsVisible:          func() *bool { b := true; return &b }(),
					BaseModel: models.BaseModel{
						CreatedAt: pastTime,
						UpdatedAt: pastTime,
					},
				},
				ParticipantCount: 15,
				TaskCount:        6,
				IsParticipant:    true,
				HasPendingReg:    false,
			},
		}

		cr.EXPECT().GetPastContestsWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(pastContestsWithStats, nil).Times(1)

		result, err := cs.GetPastContests(tx, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Past Contest 1", result[0].Name)
		assert.Equal(t, int64(15), result[0].ParticipantCount)
		assert.Equal(t, int64(6), result[0].TaskCount)
		assert.Equal(t, "registered", result[0].RegistrationStatus)
	})

	t.Run("repository error", func(t *testing.T) {
		cr.EXPECT().GetPastContestsWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(nil, gorm.ErrInvalidDB).Times(1)

		result, err := cs.GetPastContests(tx, currentUser, queryParams)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})
}

func TestContestService_GetUpcomingContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr)
	tx := &gorm.DB{}

	currentUser := schemas.User{
		ID:   1,
		Role: types.UserRoleStudent,
	}

	queryParams := map[string]any{
		"limit":  10,
		"offset": 0,
		"sort":   "created_at:desc",
	}

	t.Run("successful GetUpcomingContests", func(t *testing.T) {
		now := time.Now()
		futureTime := now.Add(24 * time.Hour)
		upcomingContestsWithStats := []models.ContestWithStats{
			{
				Contest: models.Contest{
					ID:                 1,
					Name:               "Upcoming Contest 1",
					Description:        "Description 1",
					CreatedBy:          2,
					StartAt:            &futureTime,
					EndAt:              nil,
					IsRegistrationOpen: func() *bool { b := true; return &b }(),
					IsSubmissionOpen:   func() *bool { b := false; return &b }(),
					IsVisible:          func() *bool { b := true; return &b }(),
					BaseModel: models.BaseModel{
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				ParticipantCount: 3,
				TaskCount:        2,
				IsParticipant:    false,
				HasPendingReg:    true,
			},
		}

		cr.EXPECT().GetUpcomingContestsWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(upcomingContestsWithStats, nil).Times(1)

		result, err := cs.GetUpcomingContests(tx, currentUser, queryParams)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Upcoming Contest 1", result[0].Name)
		assert.Equal(t, int64(3), result[0].ParticipantCount)
		assert.Equal(t, int64(2), result[0].TaskCount)
		assert.Equal(t, "awaitingApproval", result[0].RegistrationStatus)
	})

	t.Run("repository error", func(t *testing.T) {
		cr.EXPECT().GetUpcomingContestsWithStats(tx, currentUser.ID, 0, 10, "created_at:desc").Return(nil, gorm.ErrInvalidDB).Times(1)

		result, err := cs.GetUpcomingContests(tx, currentUser, queryParams)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, gorm.ErrInvalidDB, err)
	})
}
