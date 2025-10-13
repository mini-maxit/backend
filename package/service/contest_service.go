package service

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ContestService interface {
	Create(tx *gorm.DB, currentUser schemas.User, contest *schemas.CreateContest) (int64, error)
	GetContest(tx *gorm.DB, currentUser schemas.User, contestId int64) (*schemas.ContestDetailed, error)
	GetAll(tx *gorm.DB, currentUser schemas.User, queryParams map[string]interface{}) ([]schemas.Contest, error)
	EditContest(tx *gorm.DB, currentUser schemas.User, contestId int64, updateInfo *schemas.EditContest) error
	DeleteContest(tx *gorm.DB, currentUser schemas.User, contestId int64) error
	AssignTasksToContest(tx *gorm.DB, currentUser schemas.User, contestId int64, taskIds []int64) error
	UnAssignTasksFromContest(tx *gorm.DB, currentUser schemas.User, contestId int64, taskIds []int64) error
}

type contestService struct {
	contestRepository repository.ContestRepository
	taskRepository    repository.TaskRepository
	userRepository    repository.UserRepository
	logger            *zap.SugaredLogger
}

func (cs *contestService) Create(tx *gorm.DB, currentUser schemas.User, contest *schemas.CreateContest) (int64, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		cs.logger.Errorf("Error validating user role: %v", err.Error())
		return 0, err
	}

	// Check if contest with same title exists
	_, err = cs.contestRepository.GetContestByTitle(tx, contest.Title)
	if err != nil && err != gorm.ErrRecordNotFound {
		cs.logger.Errorf("Error getting contest by title: %v", err.Error())
		return 0, err
	} else if err == nil {
		return 0, errors.ErrContestExists
	}

	// Validate that end time is after start time
	if contest.EndTime.Before(contest.StartTime) {
		return 0, errors.ErrInvalidTimeRange
	}

	author, err := cs.userRepository.GetUser(tx, currentUser.Id)
	if err != nil {
		cs.logger.Errorf("Error getting user: %v", err.Error())
		return 0, err
	}

	model := models.Contest{
		Title:       contest.Title,
		Description: contest.Description,
		StartTime:   contest.StartTime,
		EndTime:     contest.EndTime,
		CreatedBy:   author.Id,
	}

	contestId, err := cs.contestRepository.Create(tx, &model)
	if err != nil {
		cs.logger.Errorf("Error creating contest: %v", err.Error())
		return 0, err
	}

	return contestId, nil
}

func (cs *contestService) GetContest(tx *gorm.DB, currentUser schemas.User, contestId int64) (*schemas.ContestDetailed, error) {
	contest, err := cs.contestRepository.GetContest(tx, contestId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrContestNotFound
		}
		cs.logger.Errorf("Error getting contest: %v", err.Error())
		return nil, err
	}

	tasks := make([]schemas.Task, 0, len(contest.Tasks))
	for _, task := range contest.Tasks {
		tasks = append(tasks, schemas.Task{
			Id:        task.Id,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
			CreatedAt: task.CreatedAt,
		})
	}

	return &schemas.ContestDetailed{
		Id:            contest.Id,
		Title:         contest.Title,
		Description:   contest.Description,
		StartTime:     contest.StartTime,
		EndTime:       contest.EndTime,
		CreatedBy:     contest.CreatedBy,
		CreatedByName: contest.Author.Name + " " + contest.Author.Surname,
		CreatedAt:     contest.CreatedAt,
		UpdatedAt:     contest.UpdatedAt,
		Tasks:         tasks,
	}, nil
}

func (cs *contestService) GetAll(tx *gorm.DB, currentUser schemas.User, queryParams map[string]interface{}) ([]schemas.Contest, error) {
	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "created_at:desc"
	}

	contests, err := cs.contestRepository.GetAllContests(tx, int(limit), int(offset), sort)
	if err != nil {
		cs.logger.Errorf("Error getting all contests: %v", err.Error())
		return nil, err
	}

	result := make([]schemas.Contest, 0, len(contests))
	for _, contest := range contests {
		result = append(result, schemas.Contest{
			Id:          contest.Id,
			Title:       contest.Title,
			Description: contest.Description,
			StartTime:   contest.StartTime,
			EndTime:     contest.EndTime,
			CreatedBy:   contest.CreatedBy,
			CreatedAt:   contest.CreatedAt,
			UpdatedAt:   contest.UpdatedAt,
		})
	}

	return result, nil
}

func (cs *contestService) EditContest(tx *gorm.DB, currentUser schemas.User, contestId int64, updateInfo *schemas.EditContest) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		cs.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	contest, err := cs.contestRepository.GetContest(tx, contestId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrContestNotFound
		}
		cs.logger.Errorf("Error getting contest: %v", err.Error())
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.Id {
		return errors.ErrNotAuthorized
	}

	model := models.Contest{}

	if updateInfo.Title != nil {
		// Check if new title conflicts with existing contests
		existingContest, err := cs.contestRepository.GetContestByTitle(tx, *updateInfo.Title)
		if err != nil && err != gorm.ErrRecordNotFound {
			cs.logger.Errorf("Error checking contest title: %v", err.Error())
			return err
		}
		if existingContest != nil && existingContest.Id != contestId {
			return errors.ErrContestExists
		}
		model.Title = *updateInfo.Title
	}

	if updateInfo.Description != nil {
		model.Description = *updateInfo.Description
	}

	if updateInfo.StartTime != nil {
		model.StartTime = *updateInfo.StartTime
	}

	if updateInfo.EndTime != nil {
		model.EndTime = *updateInfo.EndTime
	}

	// Validate time range if both times are being updated or if one is updated, check against existing
	startTime := contest.StartTime
	endTime := contest.EndTime
	if updateInfo.StartTime != nil {
		startTime = *updateInfo.StartTime
	}
	if updateInfo.EndTime != nil {
		endTime = *updateInfo.EndTime
	}
	if endTime.Before(startTime) {
		return errors.ErrInvalidTimeRange
	}

	err = cs.contestRepository.EditContest(tx, contestId, &model)
	if err != nil {
		cs.logger.Errorf("Error editing contest: %v", err.Error())
		return err
	}

	return nil
}

func (cs *contestService) DeleteContest(tx *gorm.DB, currentUser schemas.User, contestId int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		cs.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	contest, err := cs.contestRepository.GetContest(tx, contestId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrContestNotFound
		}
		cs.logger.Errorf("Error getting contest: %v", err.Error())
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.Id {
		return errors.ErrNotAuthorized
	}

	err = cs.contestRepository.DeleteContest(tx, contestId)
	if err != nil {
		cs.logger.Errorf("Error deleting contest: %v", err.Error())
		return err
	}

	return nil
}

func (cs *contestService) AssignTasksToContest(tx *gorm.DB, currentUser schemas.User, contestId int64, taskIds []int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		cs.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	contest, err := cs.contestRepository.GetContest(tx, contestId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrContestNotFound
		}
		cs.logger.Errorf("Error getting contest: %v", err.Error())
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.Id {
		return errors.ErrNotAuthorized
	}

	for _, taskId := range taskIds {
		_, err := cs.taskRepository.GetTask(tx, taskId)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				cs.logger.Errorf("Task not found: %v", taskId)
				return errors.ErrTaskNotFound
			}
			cs.logger.Errorf("Error getting task: %v", err.Error())
			return err
		}

		isAssigned, err := cs.contestRepository.IsTaskAssignedToContest(tx, contestId, taskId)
		if err != nil {
			cs.logger.Errorf("Error checking if task is assigned to contest: %v", err.Error())
			return err
		}

		if !isAssigned {
			err = cs.contestRepository.AssignTaskToContest(tx, contestId, taskId)
			if err != nil {
				cs.logger.Errorf("Error assigning task to contest: %v", err.Error())
				return err
			}
		}
	}

	return nil
}

func (cs *contestService) UnAssignTasksFromContest(tx *gorm.DB, currentUser schemas.User, contestId int64, taskIds []int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		cs.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	contest, err := cs.contestRepository.GetContest(tx, contestId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrContestNotFound
		}
		cs.logger.Errorf("Error getting contest: %v", err.Error())
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.Id {
		return errors.ErrNotAuthorized
	}

	for _, taskId := range taskIds {
		err = cs.contestRepository.UnAssignTaskFromContest(tx, contestId, taskId)
		if err != nil {
			cs.logger.Errorf("Error unassigning task from contest: %v", err.Error())
			return err
		}
	}

	return nil
}

func NewContestService(contestRepository repository.ContestRepository, taskRepository repository.TaskRepository, userRepository repository.UserRepository) ContestService {
	return &contestService{
		contestRepository: contestRepository,
		taskRepository:    taskRepository,
		userRepository:    userRepository,
		logger:            utils.NewNamedLogger("contest_service"),
	}
}
