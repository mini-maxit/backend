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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCanUserAccess(t *testing.T) {
	resourceType := types.ResourceTypeTask
	resourceID := int64(1)

	t.Run("Admin user has access", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		adminUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		err := acs.CanUserAccess(db, resourceType, resourceID, adminUser, types.PermissionEdit)
		require.NoError(t, err)
	})

	t.Run("User with owner permission has access to edit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 2, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.PermissionOwner, nil).Times(1)
		err := acs.CanUserAccess(db, resourceType, resourceID, user, types.PermissionEdit)
		require.NoError(t, err)
	})

	t.Run("User with manage permission has access to edit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 2, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.PermissionManage, nil).Times(1)
		err := acs.CanUserAccess(db, resourceType, resourceID, user, types.PermissionEdit)
		require.NoError(t, err)
	})

	t.Run("User with edit permission has access to edit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 2, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.PermissionEdit, nil).Times(1)
		err := acs.CanUserAccess(db, resourceType, resourceID, user, types.PermissionEdit)
		require.NoError(t, err)
	})

	t.Run("User with edit permission denied manage access", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 2, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.PermissionEdit, nil).Times(1)
		err := acs.CanUserAccess(db, resourceType, resourceID, user, types.PermissionManage)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("User not found returns forbidden", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 3, Role: types.UserRoleStudent}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.Permission(""), gorm.ErrRecordNotFound).Times(1)
		err := acs.CanUserAccess(db, resourceType, resourceID, user, types.PermissionEdit)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("Database error returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 4, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.Permission(""), assert.AnError).Times(1)
		err := acs.CanUserAccess(db, resourceType, resourceID, user, types.PermissionEdit)
		require.Error(t, err)
		require.NotErrorIs(t, err, errors.ErrForbidden)
	})
}

func TestAddCollaborator(t *testing.T) {
	resourceType := types.ResourceTypeTask
	resourceID := int64(1)
	collaboratorID := int64(10)

	t.Run("Cannot assign owner permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionOwner)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("User without manage permission cannot add collaborator", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 2, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.PermissionEdit, nil).Times(1)
		err := acs.AddCollaborator(db, user, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("Resource not found returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		tr.EXPECT().Get(db, resourceID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Contest resource not found returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		contestResourceType := types.ResourceTypeContest
		cr.EXPECT().Get(db, resourceID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := acs.AddCollaborator(db, currentUser, contestResourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Access already exists returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(&models.AccessControl{}, nil).Times(1)
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.ErrorIs(t, err, errors.ErrAccessAlreadyExists)
	})

	t.Run("Target user not found returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Get(db, collaboratorID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Success adds collaborator", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Get(db, collaboratorID).Return(&models.User{ID: collaboratorID, Role: types.UserRoleTeacher}, nil).Times(1)
		acr.EXPECT().AddAccess(db, gomock.Any()).Return(nil).Times(1)
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.NoError(t, err)
	})

	t.Run("Duplicate key error returns access already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Get(db, collaboratorID).Return(&models.User{ID: collaboratorID, Role: types.UserRoleTeacher}, nil).Times(1)
		acr.EXPECT().AddAccess(db, gomock.Any()).Return(gorm.ErrDuplicatedKey).Times(1)
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.ErrorIs(t, err, errors.ErrAccessAlreadyExists)
	})

	t.Run("Success with manage permission user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 5, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.PermissionManage, nil).Times(1)
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Get(db, collaboratorID).Return(&models.User{ID: collaboratorID, Role: types.UserRoleTeacher}, nil).Times(1)
		acr.EXPECT().AddAccess(db, gomock.Any()).Return(nil).Times(1)
		err := acs.AddCollaborator(db, user, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.NoError(t, err)
	})
}

func TestGetCollaborators(t *testing.T) {
	resourceType := types.ResourceTypeTask
	resourceID := int64(1)

	t.Run("User without edit permission cannot get collaborators", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 2, Role: types.UserRoleStudent}
		// GetCollaborators checks resource existence BEFORE checking permissions
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.Permission(""), gorm.ErrRecordNotFound).Times(1)
		collaborators, err := acs.GetCollaborators(db, user, resourceType, resourceID)
		require.ErrorIs(t, err, errors.ErrForbidden)
		assert.Nil(t, collaborators)
	})

	t.Run("Resource not found returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		// checkResourceExists checks the boolean and returns ErrNotFound if !exists
		tr.EXPECT().Get(db, resourceID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		// GetResourceAccess should NOT be called because the resource doesn't exist
		collaborators, err := acs.GetCollaborators(db, user, resourceType, resourceID)
		require.ErrorIs(t, err, errors.ErrNotFound)
		assert.Nil(t, collaborators)
	})

	t.Run("Success returns collaborators", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		now := time.Now()
		accesses := []models.AccessControl{
			{
				ResourceType: resourceType,
				ResourceID:   resourceID,
				UserID:       10,
				Permission:   types.PermissionEdit,
				BaseModel:    models.BaseModel{CreatedAt: now},
				User:         models.User{ID: 10, Name: "Test User", Email: "test@example.com"},
			},
			{
				ResourceType: resourceType,
				ResourceID:   resourceID,
				UserID:       11,
				Permission:   types.PermissionManage,
				BaseModel:    models.BaseModel{CreatedAt: now},
				User:         models.User{ID: 11, Name: "Another User", Email: "another@example.com"},
			},
		}
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetResourceAccess(db, resourceType, resourceID).Return(accesses, nil).Times(1)
		collaborators, err := acs.GetCollaborators(db, user, resourceType, resourceID)
		require.NoError(t, err)
		assert.Len(t, collaborators, 2)
		assert.Equal(t, int64(10), collaborators[0].UserID)
		assert.Equal(t, "Test User", collaborators[0].UserName)
		assert.Equal(t, "test@example.com", collaborators[0].UserEmail)
		assert.Equal(t, types.PermissionEdit, collaborators[0].Permission)
	})

	t.Run("Empty collaborators list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetResourceAccess(db, resourceType, resourceID).Return([]models.AccessControl{}, nil).Times(1)
		collaborators, err := acs.GetCollaborators(db, user, resourceType, resourceID)
		require.NoError(t, err)
		assert.Empty(t, collaborators)
	})

	t.Run("Database error returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		tr.EXPECT().Get(db, resourceID).Return(&models.Task{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetResourceAccess(db, resourceType, resourceID).Return(nil, assert.AnError).Times(1)
		collaborators, err := acs.GetCollaborators(db, user, resourceType, resourceID)
		require.Error(t, err)
		assert.Nil(t, collaborators)
	})
}

func TestUpdateCollaborator(t *testing.T) {
	resourceType := types.ResourceTypeTask
	resourceID := int64(1)
	collaboratorID := int64(10)

	t.Run("Cannot assign owner permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		err := acs.UpdateCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionOwner)
		require.ErrorIs(t, err, errors.ErrCannotAssignOwner)
	})

	t.Run("User without manage permission cannot update collaborator", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 2, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.PermissionEdit, nil).Times(1)
		err := acs.UpdateCollaborator(db, user, resourceType, resourceID, collaboratorID, types.PermissionManage)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("Collaborator not found returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := acs.UpdateCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionManage)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Cannot update owner permission entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(&models.AccessControl{
			Permission: types.PermissionOwner,
		}, nil).Times(1)
		err := acs.UpdateCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionManage)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("Success updates collaborator permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(&models.AccessControl{
			Permission: types.PermissionEdit,
		}, nil).Times(1)
		acr.EXPECT().UpdatePermission(db, resourceType, resourceID, collaboratorID, types.PermissionManage).Return(nil).Times(1)
		err := acs.UpdateCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionManage)
		require.NoError(t, err)
	})

	t.Run("Database error on update returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(&models.AccessControl{
			Permission: types.PermissionEdit,
		}, nil).Times(1)
		acr.EXPECT().UpdatePermission(db, resourceType, resourceID, collaboratorID, types.PermissionManage).Return(assert.AnError).Times(1)
		err := acs.UpdateCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionManage)
		require.Error(t, err)
	})
}

func TestRemoveCollaborator(t *testing.T) {
	resourceType := types.ResourceTypeTask
	resourceID := int64(1)
	collaboratorID := int64(10)

	t.Run("User without manage permission cannot remove collaborator", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		user := &schemas.User{ID: 2, Role: types.UserRoleTeacher}
		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, user.ID).Return(types.PermissionEdit, nil).Times(1)
		err := acs.RemoveCollaborator(db, user, resourceType, resourceID, collaboratorID)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("Collaborator not found returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		err := acs.RemoveCollaborator(db, currentUser, resourceType, resourceID, collaboratorID)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Cannot remove owner permission entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(&models.AccessControl{
			Permission: types.PermissionOwner,
		}, nil).Times(1)
		err := acs.RemoveCollaborator(db, currentUser, resourceType, resourceID, collaboratorID)
		require.ErrorIs(t, err, errors.ErrForbidden)
	})

	t.Run("Success removes collaborator", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(&models.AccessControl{
			Permission: types.PermissionEdit,
		}, nil).Times(1)
		acr.EXPECT().RemoveAccess(db, resourceType, resourceID, collaboratorID).Return(nil).Times(1)
		err := acs.RemoveCollaborator(db, currentUser, resourceType, resourceID, collaboratorID)
		require.NoError(t, err)
	})

	t.Run("Database error on remove returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(&models.AccessControl{
			Permission: types.PermissionEdit,
		}, nil).Times(1)
		acr.EXPECT().RemoveAccess(db, resourceType, resourceID, collaboratorID).Return(assert.AnError).Times(1)
		err := acs.RemoveCollaborator(db, currentUser, resourceType, resourceID, collaboratorID)
		require.Error(t, err)
	})
}

func TestGetUserPermission(t *testing.T) {
	resourceType := types.ResourceTypeTask
	resourceID := int64(1)
	userID := int64(10)

	t.Run("Success returns permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, userID).Return(types.PermissionEdit, nil).Times(1)
		permission, err := acs.GetUserPermission(db, resourceType, resourceID, userID)
		require.NoError(t, err)
		assert.Equal(t, types.PermissionEdit, permission)
	})

	t.Run("User not found returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, userID).Return(types.Permission(""), gorm.ErrRecordNotFound).Times(1)
		permission, err := acs.GetUserPermission(db, resourceType, resourceID, userID)
		require.ErrorIs(t, err, errors.ErrNotFound)
		assert.Equal(t, types.Permission(""), permission)
	})

	t.Run("Database error returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		acr.EXPECT().GetUserPermission(db, resourceType, resourceID, userID).Return(types.Permission(""), assert.AnError).Times(1)
		permission, err := acs.GetUserPermission(db, resourceType, resourceID, userID)
		require.Error(t, err)
		assert.Equal(t, types.Permission(""), permission)
	})
}

func TestGrantOwnerAccess(t *testing.T) {
	resourceType := types.ResourceTypeTask
	resourceID := int64(1)
	ownerID := int64(10)

	t.Run("Success grants owner access", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		acr.EXPECT().AddAccess(db, gomock.Any()).DoAndReturn(func(_ interface{}, access *models.AccessControl) error {
			assert.Equal(t, resourceType, access.ResourceType)
			assert.Equal(t, resourceID, access.ResourceID)
			assert.Equal(t, ownerID, access.UserID)
			assert.Equal(t, types.PermissionOwner, access.Permission)
			return nil
		}).Times(1)
		err := acs.GrantOwnerAccess(db, resourceType, resourceID, ownerID)
		require.NoError(t, err)
	})

	t.Run("Error on add access does not fail the operation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		acr.EXPECT().AddAccess(db, gomock.Any()).Return(assert.AnError).Times(1)
		err := acs.GrantOwnerAccess(db, resourceType, resourceID, ownerID)
		require.NoError(t, err)
	})

	t.Run("Grants owner access for contest resource type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		contestResourceType := types.ResourceTypeContest
		acr.EXPECT().AddAccess(db, gomock.Any()).DoAndReturn(func(_ interface{}, access *models.AccessControl) error {
			assert.Equal(t, contestResourceType, access.ResourceType)
			assert.Equal(t, resourceID, access.ResourceID)
			assert.Equal(t, ownerID, access.UserID)
			assert.Equal(t, types.PermissionOwner, access.Permission)
			return nil
		}).Times(1)
		err := acs.GrantOwnerAccess(db, contestResourceType, resourceID, ownerID)
		require.NoError(t, err)
	})
}

func TestCheckResourceExists(t *testing.T) {
	resourceID := int64(1)
	collaboratorID := int64(10)

	t.Run("Invalid resource type returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		invalidResourceType := types.ResourceType("invalid")
		err := acs.AddCollaborator(db, currentUser, invalidResourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.ErrorIs(t, err, errors.ErrInvalidData)
	})

	t.Run("Task repository error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		resourceType := types.ResourceTypeTask
		tr.EXPECT().Get(db, resourceID).Return(nil, assert.AnError).Times(1)
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.Error(t, err)
		require.NotErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Contest repository error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		resourceType := types.ResourceTypeContest
		cr.EXPECT().Get(db, resourceID).Return(nil, assert.AnError).Times(1)
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionEdit)
		require.Error(t, err)
		require.NotErrorIs(t, err, errors.ErrNotFound)
	})
}

func TestAddCollaboratorWithContestResource(t *testing.T) {
	resourceType := types.ResourceTypeContest
	resourceID := int64(1)
	collaboratorID := int64(10)

	t.Run("Success adds collaborator to contest", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		// Admin user - CanUserAccess returns early without calling GetUserPermission
		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		cr.EXPECT().Get(db, resourceID).Return(&models.Contest{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetAccess(db, resourceType, resourceID, collaboratorID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Get(db, collaboratorID).Return(&models.User{ID: collaboratorID, Role: types.UserRoleTeacher}, nil).Times(1)
		acr.EXPECT().AddAccess(db, gomock.Any()).Return(nil).Times(1)
		err := acs.AddCollaborator(db, currentUser, resourceType, resourceID, collaboratorID, types.PermissionManage)
		require.NoError(t, err)
	})
}

func TestGetCollaboratorsWithContestResource(t *testing.T) {
	resourceType := types.ResourceTypeContest
	resourceID := int64(1)

	t.Run("Success returns contest collaborators", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := &testutils.MockDatabase{}
		acr := mock_repository.NewMockAccessControlRepository(ctrl)
		ur := mock_repository.NewMockUserRepository(ctrl)
		tr := mock_repository.NewMockTaskRepository(ctrl)
		cr := mock_repository.NewMockContestRepository(ctrl)
		gr := mock_repository.NewMockGroupRepository(ctrl)
		acs := service.NewAccessControlService(acr, ur, tr, cr, gr)

		currentUser := &schemas.User{ID: 1, Role: types.UserRoleAdmin}
		now := time.Now()
		accesses := []models.AccessControl{
			{
				ResourceType: resourceType,
				ResourceID:   resourceID,
				UserID:       10,
				Permission:   types.PermissionManage,
				BaseModel:    models.BaseModel{CreatedAt: now},
				User:         models.User{ID: 10, Name: "Contest Manager", Email: "manager@example.com"},
			},
		}
		cr.EXPECT().Get(db, resourceID).Return(&models.Contest{ID: resourceID}, nil).Times(1)
		acr.EXPECT().GetResourceAccess(db, resourceType, resourceID).Return(accesses, nil).Times(1)
		collaborators, err := acs.GetCollaborators(db, currentUser, resourceType, resourceID)
		require.NoError(t, err)
		assert.Len(t, collaborators, 1)
		assert.Equal(t, int64(10), collaborators[0].UserID)
		assert.Equal(t, types.PermissionManage, collaborators[0].Permission)
	})
}
