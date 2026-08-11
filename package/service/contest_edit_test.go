package service_test

import (
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func strPtr(s string) *string { return &s }

func newContestServiceForEditTest(ctrl *gomock.Controller) (*mock_repository.MockContestRepository, *mock_service.MockAccessControlService, service.ContestService) {
	cr := mock_repository.NewMockContestRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	sr := mock_repository.NewMockSubmissionRepository(ctrl)
	tr := mock_repository.NewMockTaskRepository(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	acs := mock_service.NewMockAccessControlService(ctrl)
	gr := mock_repository.NewMockGroupRepository(ctrl)
	cs := service.NewContestService(cr, ur, sr, tr, gr, acs, ts)
	return cr, acs, cs
}

func TestContestServiceEdit_ClearEndAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr, acs, cs := newContestServiceForEditTest(ctrl)

	db := &testutils.MockDatabase{}
	userID := int64(1)
	currentUser := &schemas.User{ID: userID, Role: types.UserRoleAdmin}
	contestID := int64(5)

	// Contest currently has an end time
	endAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	contest := &models.Contest{ID: contestID, StartAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), EndAt: &endAt}

	// Request explicitly sets endAt to null (clear it)
	editInfo := &schemas.EditContest{
		EndAt: schemas.OptionalTime{Set: true, Value: nil},
	}

	acs.EXPECT().CanUserAccess(db, types.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
	// Need current contest because EndAt is being cleared (StartAt also nil -> fetch current)
	cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(2)

	cr.EXPECT().EditWithStats(db, contestID, gomock.Any()).DoAndReturn(
		func(_ interface{}, id int64, updates map[string]any) (*models.ContestWithStats, error) {
			// end_at must be present in updates with nil value
			assert.Contains(t, updates, "end_at")
			assert.Nil(t, updates["end_at"])
			return &models.ContestWithStats{Contest: *contest}, nil
		},
	).Times(1)

	_, err := cs.Edit(db, currentUser, contestID, editInfo)
	require.NoError(t, err)
}

func TestContestServiceEdit_KeepEndAtWhenAbsent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr, acs, cs := newContestServiceForEditTest(ctrl)

	db := &testutils.MockDatabase{}
	currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
	contestID := int64(5)

	// Edit only the name; EndAt absent (Set=false) -> not in update map
	editInfo := &schemas.EditContest{
		Name: strPtr("New Name"),
	}

	acs.EXPECT().CanUserAccess(db, types.ResourceTypeContest, contestID, currentUser, types.PermissionEdit).Return(nil).Times(1)
	// StartAt and EndAt both nil -> need current contest
	endAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	contest := &models.Contest{ID: contestID, StartAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), EndAt: &endAt}
	cr.EXPECT().Get(db, contestID).Return(contest, nil).Times(2)

	cr.EXPECT().EditWithStats(db, contestID, gomock.Any()).DoAndReturn(
		func(_ interface{}, id int64, updates map[string]any) (*models.ContestWithStats, error) {
			assert.Contains(t, updates, "name")
			assert.NotContains(t, updates, "end_at")
			return &models.ContestWithStats{Contest: *contest}, nil
		},
	).Times(1)

	_, err := cs.Edit(db, currentUser, contestID, editInfo)
	require.NoError(t, err)
}
