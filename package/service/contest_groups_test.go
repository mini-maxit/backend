package service_test

import (
	"testing"

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
		cr.EXPECT().AddGroupToContest(db, contestID, groupID).Return(nil).Times(1)

		err := cs.AddGroupToContest(db, currentUser, contestID, groupID)
		require.NoError(t, err)
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
		cr.EXPECT().RemoveGroupFromContest(db, contestID, groupID).Return(nil).Times(1)

		err := cs.RemoveGroupFromContest(db, currentUser, contestID, groupID)
		require.NoError(t, err)
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
		assignableGroups := []models.Group{
			{ID: 3, Name: "Group 3", CreatedBy: 10},
			{ID: 4, Name: "Group 4", CreatedBy: 10},
		}

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID, CreatedBy: 10}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().GetContestGroups(db, contestID).Return(assignedGroups, nil).Times(1)
		cr.EXPECT().GetAssignableGroups(db, contestID).Return(assignableGroups, nil).Times(1)

		result, err := cs.GetContestGroups(db, currentUser, contestID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Assigned, 2)
		assert.Len(t, result.Assignable, 2)
		assert.Equal(t, int64(1), result.Assigned[0].ID)
		assert.Equal(t, "Group 1", result.Assigned[0].Name)
		assert.Equal(t, int64(3), result.Assignable[0].ID)
		assert.Equal(t, "Group 3", result.Assignable[0].Name)
	})

	t.Run("success with empty groups", func(t *testing.T) {
		currentUser := &schemas.User{ID: 10, Role: types.UserRoleAdmin}
		contestID := int64(1)

		cr.EXPECT().Get(db, contestID).Return(&models.Contest{ID: contestID}, nil).Times(1)
		acs.EXPECT().CanUserAccess(db, models.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
		cr.EXPECT().GetContestGroups(db, contestID).Return([]models.Group{}, nil).Times(1)
		cr.EXPECT().GetAssignableGroups(db, contestID).Return([]models.Group{}, nil).Times(1)

		result, err := cs.GetContestGroups(db, currentUser, contestID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Assigned, 0)
		assert.Len(t, result.Assignable, 0)
	})
}
