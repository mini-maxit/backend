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

func newRegistrationTestService(t *testing.T) (service.ContestService, *mock_repository.MockContestRepository, *mock_service.MockAccessControlService) {
	ctrl := gomock.NewController(t)
	cr := mock_repository.NewMockContestRepository(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	cs := service.NewContestService(
		cr,
		mock_repository.NewMockUserRepository(ctrl),
		mock_repository.NewMockSubmissionRepository(ctrl),
		mock_repository.NewMockTaskRepository(ctrl),
		mock_repository.NewMockGroupRepository(ctrl),
		acs,
		mock_service.NewMockTaskService(ctrl),
	)
	return cs, cr, acs
}

func registrationStatusPtr(s types.RegistrationRequestStatus) *types.RegistrationRequestStatus {
	return &s
}

func TestContestService_GetDetailed_PublicContestVisibleToEveryone(t *testing.T) {
	cs, cr, _ := newRegistrationTestService(t)
	db := &testutils.MockDatabase{}
	student := &schemas.User{ID: 1, Role: types.UserRoleStudent}
	contest := &repository.ContestDetailed{Contest: models.Contest{ID: 10, Name: testContestName, IsPublic: true}}

	// No access-control or participant lookups: public contests are viewable by anyone.
	cr.EXPECT().GetDetailed(db, int64(10)).Return(contest, nil)

	result, err := cs.GetDetailed(db, student, 10)

	require.NoError(t, err)
	assert.Equal(t, int64(10), result.ID)
}

func TestContestService_GetPublicContests_RegistrationStatus(t *testing.T) {
	cs, cr, _ := newRegistrationTestService(t)
	db := &testutils.MockDatabase{}
	student := &schemas.User{ID: 1, Role: types.UserRoleStudent}
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)

	contests := []models.PublicContest{
		{Contest: models.Contest{ID: 1, EndAt: &future}, IsParticipant: true},
		{Contest: models.Contest{ID: 2}, RequestStatus: registrationStatusPtr(types.RegistrationRequestStatusPending)},
		{Contest: models.Contest{ID: 3}, RequestStatus: registrationStatusPtr(types.RegistrationRequestStatusRejected)},
		{Contest: models.Contest{ID: 4, EndAt: &past}},
		{Contest: models.Contest{ID: 5, EndAt: &future}, CreatorName: testCreatorName, ParticipantCount: 3, TaskCount: 2},
	}
	cr.EXPECT().GetPublicContests(db, student.ID, types.ContestStatusOngoing, 0, 10, "created_at:desc").Return(contests, int64(5), nil)

	result, err := cs.GetPublicContests(db, student, types.ContestStatusOngoing, schemas.PaginationParams{Limit: 10})

	require.NoError(t, err)
	require.Len(t, result.Items, 5)
	assert.Equal(t, types.RegistrationStatusRegistered, result.Items[0].RegistrationStatus)
	assert.Equal(t, types.RegistrationStatusAwaitingApproval, result.Items[1].RegistrationStatus)
	assert.Equal(t, types.RegistrationStatusRejected, result.Items[2].RegistrationStatus)
	assert.Equal(t, types.RegistrationStatusRegistrationClosed, result.Items[3].RegistrationStatus)
	assert.Equal(t, types.RegistrationStatusCanRegister, result.Items[4].RegistrationStatus)
	assert.Equal(t, testCreatorName, result.Items[4].CreatorName)
	assert.Equal(t, int64(3), result.Items[4].ParticipantCount)
	assert.Equal(t, 5, result.Pagination.TotalItems)
}

func TestContestService_RegisterForContest(t *testing.T) {
	db := &testutils.MockDatabase{}
	student := &schemas.User{ID: 1, Role: types.UserRoleStudent}
	contestID := int64(10)
	past := time.Now().Add(-time.Hour)

	t.Run("creates pending request", func(t *testing.T) {
		cs, cr, _ := newRegistrationTestService(t)
		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, IsPublic: true}, nil)
		cr.EXPECT().IsUserParticipant(db, contestID, student.ID).Return(false, nil)
		cr.EXPECT().GetRegistrationRequest(db, contestID, student.ID).Return(nil, gorm.ErrRecordNotFound)
		cr.EXPECT().CreateRegistrationRequest(db, &models.ContestRegistrationRequest{
			ContestID: contestID, UserID: student.ID, Status: types.RegistrationRequestStatusPending,
		}).Return(nil)

		require.NoError(t, cs.RegisterForContest(db, student, contestID))
	})

	t.Run("missing contest is not found", func(t *testing.T) {
		cs, cr, _ := newRegistrationTestService(t)
		cr.EXPECT().Get(db, contestID).Return(nil, gorm.ErrRecordNotFound)

		require.ErrorIs(t, cs.RegisterForContest(db, student, contestID), errors.ErrNotFound)
	})

	t.Run("private contest is reported as not found", func(t *testing.T) {
		cs, cr, _ := newRegistrationTestService(t)
		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil)

		require.ErrorIs(t, cs.RegisterForContest(db, student, contestID), errors.ErrNotFound)
	})

	t.Run("ended contest", func(t *testing.T) {
		cs, cr, _ := newRegistrationTestService(t)
		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, IsPublic: true, EndAt: &past}, nil)

		require.ErrorIs(t, cs.RegisterForContest(db, student, contestID), errors.ErrContestEnded)
	})

	t.Run("already participant", func(t *testing.T) {
		cs, cr, _ := newRegistrationTestService(t)
		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, IsPublic: true}, nil)
		cr.EXPECT().IsUserParticipant(db, contestID, student.ID).Return(true, nil)

		require.ErrorIs(t, cs.RegisterForContest(db, student, contestID), errors.ErrAlreadyParticipant)
	})

	for _, status := range []types.RegistrationRequestStatus{types.RegistrationRequestStatusPending, types.RegistrationRequestStatusRejected} {
		t.Run("existing "+string(status)+" request", func(t *testing.T) {
			cs, cr, _ := newRegistrationTestService(t)
			cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, IsPublic: true}, nil)
			cr.EXPECT().IsUserParticipant(db, contestID, student.ID).Return(false, nil)
			cr.EXPECT().GetRegistrationRequest(db, contestID, student.ID).Return(&models.ContestRegistrationRequest{Status: status}, nil)

			require.ErrorIs(t, cs.RegisterForContest(db, student, contestID), errors.ErrAlreadyRegistered)
		})
	}
}

func TestContestService_ResolveRegistrationRequests(t *testing.T) {
	db := &testutils.MockDatabase{}
	teacher := &schemas.User{ID: 2, Role: types.UserRoleTeacher}
	contestID, userID := int64(10), int64(5)
	pending := &models.ContestRegistrationRequest{ContestID: contestID, UserID: userID, Status: types.RegistrationRequestStatusPending}

	expectManager := func(cr *mock_repository.MockContestRepository, acs *mock_service.MockAccessControlService, allowed bool) {
		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil)
		var err error
		if !allowed {
			err = errors.ErrForbidden
		}
		acs.EXPECT().CanUserAccess(db, types.ResourceTypeContest, contestID, teacher, types.PermissionEdit).Return(err)
	}

	t.Run("approve adds participant and removes request", func(t *testing.T) {
		cs, cr, acs := newRegistrationTestService(t)
		expectManager(cr, acs, true)
		cr.EXPECT().GetRegistrationRequest(db, contestID, userID).Return(pending, nil)
		cr.EXPECT().DeleteRegistrationRequest(db, contestID, userID).Return(nil)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(false, nil)
		cr.EXPECT().CreateContestParticipant(db, contestID, userID).Return(nil)

		require.NoError(t, cs.ApproveRegistrationRequest(db, teacher, contestID, userID))
	})

	t.Run("approve when already participant only removes request", func(t *testing.T) {
		cs, cr, acs := newRegistrationTestService(t)
		expectManager(cr, acs, true)
		cr.EXPECT().GetRegistrationRequest(db, contestID, userID).Return(pending, nil)
		cr.EXPECT().DeleteRegistrationRequest(db, contestID, userID).Return(nil)
		cr.EXPECT().IsUserParticipant(db, contestID, userID).Return(true, nil)

		require.NoError(t, cs.ApproveRegistrationRequest(db, teacher, contestID, userID))
	})

	t.Run("reject marks request rejected", func(t *testing.T) {
		cs, cr, acs := newRegistrationTestService(t)
		expectManager(cr, acs, true)
		cr.EXPECT().GetRegistrationRequest(db, contestID, userID).Return(pending, nil)
		cr.EXPECT().UpdateRegistrationRequestStatus(db, contestID, userID, types.RegistrationRequestStatusRejected).Return(nil)

		require.NoError(t, cs.RejectRegistrationRequest(db, teacher, contestID, userID))
	})

	t.Run("no pending request", func(t *testing.T) {
		cs, cr, acs := newRegistrationTestService(t)
		expectManager(cr, acs, true)
		cr.EXPECT().GetRegistrationRequest(db, contestID, userID).Return(nil, gorm.ErrRecordNotFound)
		require.ErrorIs(t, cs.ApproveRegistrationRequest(db, teacher, contestID, userID), errors.ErrNoPendingRegistration)
	})

	t.Run("already rejected request cannot be approved", func(t *testing.T) {
		cs, cr, acs := newRegistrationTestService(t)
		expectManager(cr, acs, true)
		cr.EXPECT().GetRegistrationRequest(db, contestID, userID).Return(
			&models.ContestRegistrationRequest{Status: types.RegistrationRequestStatusRejected}, nil)
		require.ErrorIs(t, cs.ApproveRegistrationRequest(db, teacher, contestID, userID), errors.ErrNoPendingRegistration)
	})

	t.Run("non-collaborator is forbidden", func(t *testing.T) {
		cs, cr, acs := newRegistrationTestService(t)
		expectManager(cr, acs, false)
		require.ErrorIs(t, cs.RejectRegistrationRequest(db, teacher, contestID, userID), errors.ErrForbidden)
	})

	t.Run("list maps users", func(t *testing.T) {
		cs, cr, acs := newRegistrationTestService(t)
		expectManager(cr, acs, true)
		cr.EXPECT().GetPendingRegistrationRequests(db, contestID).Return([]models.ContestRegistrationRequest{
			{ContestID: contestID, UserID: userID, User: models.User{ID: userID, Name: "Ann"}},
		}, nil)

		result, err := cs.GetRegistrationRequests(db, teacher, contestID)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, userID, result[0].User.ID)
		assert.Equal(t, "Ann", result[0].User.Name)
	})
}
