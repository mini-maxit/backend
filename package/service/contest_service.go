package service

import (
	"errors"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ContestService interface {
	// Create creates a new contest
	Create(tx *gorm.DB, currentUser schemas.User, contest *schemas.CreateContest) (int64, error)
	// Get retrieves a contest by ID
	Get(tx *gorm.DB, currentUser schemas.User, contestID int64) (*schemas.Contest, error)
	// GetAll retrieves all contests with pagination
	GetAll(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.Contest, error)
	// Edit updates a contest
	Edit(tx *gorm.DB, currentUser schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.Contest, error)
	// Delete removes a contest
	Delete(tx *gorm.DB, currentUser schemas.User, contestID int64) error
}

const defaultContestSort = "created_at:desc"

type contestService struct {
	contestRepository repository.ContestRepository
	userRepository    repository.UserRepository

	logger *zap.SugaredLogger
}

func (cs *contestService) Create(tx *gorm.DB, currentUser schemas.User, contest *schemas.CreateContest) (int64, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return 0, err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return -1, err
	}
	if err := validate.Struct(contest); err != nil {
		return 0, err
	}

	model := &models.Contest{
		Name:               contest.Name,
		Description:        contest.Description,
		CreatedBy:          currentUser.ID,
		IsRegistrationOpen: contest.IsRegistrationOpen,
		IsSubmissionOpen:   contest.IsSubmissionOpen,
		IsVisible:          contest.IsVisible,
	}

	if contest.StartAt != nil {
		model.StartAt = contest.StartAt
	}
	if contest.EndAt != nil {
		model.EndAt = contest.EndAt
	}

	contestID, err := cs.contestRepository.Create(tx, model)
	if err != nil {
		return -1, err
	}

	return contestID, nil
}

func (cs *contestService) Get(tx *gorm.DB, currentUser schemas.User, contestID int64) (*schemas.Contest, error) {
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	// Check visibility and permissions
	if !*contest.IsVisible {
		if currentUser.Role == types.UserRoleStudent {
			return nil, myerrors.ErrNotAuthorized
		}
		if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.ID {
			return nil, myerrors.ErrNotAuthorized
		}
	}

	return ContestToSchema(contest), nil
}

func (cs *contestService) GetAll(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.Contest, error) {
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultContestSort
	}

	var contests []models.Contest
	var err error

	switch currentUser.Role {
	case types.UserRoleAdmin:
		contests, err = cs.contestRepository.GetAll(tx, offset, limit, sort)
		if err != nil {
			return nil, err
		}
	case types.UserRoleTeacher:
		contests, err = cs.contestRepository.GetAllForCreator(tx, currentUser.ID, offset, limit, sort)
		if err != nil {
			return nil, err
		}
	case types.UserRoleStudent:
		// Students can only see visible contests
		contests, err = cs.contestRepository.GetAll(tx, offset, limit, sort)
		if err != nil {
			return nil, err
		}
		// Filter visible contests
		visibleContests := []models.Contest{}
		for _, contest := range contests {
			if *contest.IsVisible {
				visibleContests = append(visibleContests, contest)
			}
		}
		contests = visibleContests
	default:
		return nil, myerrors.ErrNotAuthorized
	}

	result := make([]schemas.Contest, len(contests))
	for i, contest := range contests {
		result[i] = *ContestToSchema(&contest)
	}

	return result, nil
}

func (cs *contestService) Edit(tx *gorm.DB, currentUser schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.Contest, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return nil, err
	}
	if err := validate.Struct(editInfo); err != nil {
		return nil, err
	}

	model, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	// Check permissions
	if currentUser.Role == types.UserRoleTeacher && model.CreatedBy != currentUser.ID {
		return nil, myerrors.ErrNotAuthorized
	}

	cs.updateModel(model, editInfo)

	newModel, err := cs.contestRepository.Edit(tx, contestID, model)
	if err != nil {
		return nil, err
	}

	return ContestToSchema(newModel), nil
}

func (cs *contestService) Delete(tx *gorm.DB, currentUser schemas.User, contestID int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	return cs.contestRepository.Delete(tx, contestID)
}

func (cs *contestService) updateModel(model *models.Contest, editInfo *schemas.EditContest) {
	if editInfo.Name != nil {
		model.Name = *editInfo.Name
	}
	if editInfo.Description != nil {
		model.Description = *editInfo.Description
	}
	// TODO: handle when setting to nil is intended
	if editInfo.StartAt != nil {
		model.StartAt = editInfo.StartAt
	}
	// TODO: handle when setting to nil is intended
	if editInfo.EndAt != nil {
		model.EndAt = editInfo.EndAt
	}
	if editInfo.IsRegistrationOpen != nil {
		cs.logger.Infof("Setting IsRegistrationOpen to %v", *editInfo.IsRegistrationOpen)
		model.IsRegistrationOpen = editInfo.IsRegistrationOpen
	}
	if editInfo.IsSubmissionOpen != nil {
		cs.logger.Infof("Setting IsSubmissionOpen to %v", *editInfo.IsSubmissionOpen)
		model.IsSubmissionOpen = editInfo.IsSubmissionOpen
	}
	if editInfo.IsVisible != nil {
		cs.logger.Infof("Setting IsVisible to %v", *editInfo.IsVisible)
		model.IsVisible = editInfo.IsVisible
	}
}

func ContestToSchema(model *models.Contest) *schemas.Contest {
	return &schemas.Contest{
		ID:                 model.ID,
		Name:               model.Name,
		Description:        model.Description,
		CreatedBy:          model.CreatedBy,
		StartAt:            model.StartAt,
		EndAt:              model.EndAt,
		IsRegistrationOpen: *model.IsRegistrationOpen,
		IsSubmissionOpen:   *model.IsSubmissionOpen,
		IsVisible:          *model.IsVisible,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

func NewContestService(contestRepository repository.ContestRepository, userRepository repository.UserRepository) ContestService {
	return &contestService{
		contestRepository: contestRepository,
		userRepository:    userRepository,
		logger:            utils.NewNamedLogger("contest_service"),
	}
}
