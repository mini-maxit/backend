package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type ContestRepository interface {
	Create(tx *gorm.DB, contest *models.Contest) (int64, error)
	GetContest(tx *gorm.DB, contestId int64) (*models.Contest, error)
	GetAllContests(tx *gorm.DB, limit, offset int, sort string) ([]models.Contest, error)
	GetContestByTitle(tx *gorm.DB, title string) (*models.Contest, error)
	EditContest(tx *gorm.DB, contestId int64, contest *models.Contest) error
	DeleteContest(tx *gorm.DB, contestId int64) error
	AssignTaskToContest(tx *gorm.DB, contestId, taskId int64) error
	UnAssignTaskFromContest(tx *gorm.DB, contestId, taskId int64) error
	IsTaskAssignedToContest(tx *gorm.DB, contestId, taskId int64) (bool, error)
}

type contestRepository struct {
}

func (cr *contestRepository) Create(tx *gorm.DB, contest *models.Contest) (int64, error) {
	err := tx.Model(models.Contest{}).Create(&contest).Error
	if err != nil {
		return 0, err
	}
	return contest.Id, nil
}

func (cr *contestRepository) GetContest(tx *gorm.DB, contestId int64) (*models.Contest, error) {
	contest := &models.Contest{}
	err := tx.Preload("Author").Preload("Tasks").Model(&models.Contest{}).Where("id = ?", contestId).First(contest).Error
	if err != nil {
		return nil, err
	}
	return contest, nil
}

func (cr *contestRepository) GetAllContests(tx *gorm.DB, limit, offset int, sort string) ([]models.Contest, error) {
	var contests []models.Contest
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.Contest{}).Find(&contests).Error
	if err != nil {
		return nil, err
	}
	return contests, nil
}

func (cr *contestRepository) GetContestByTitle(tx *gorm.DB, title string) (*models.Contest, error) {
	contest := &models.Contest{}
	err := tx.Model(&models.Contest{}).Where("title = ?", title).First(contest).Error
	if err != nil {
		return nil, err
	}
	return contest, nil
}

func (cr *contestRepository) EditContest(tx *gorm.DB, contestId int64, contest *models.Contest) error {
	err := tx.Model(&models.Contest{}).Where("id = ?", contestId).Updates(contest).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) DeleteContest(tx *gorm.DB, contestId int64) error {
	err := tx.Model(&models.Contest{}).Where("id = ?", contestId).Delete(&models.Contest{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) AssignTaskToContest(tx *gorm.DB, contestId, taskId int64) error {
	contestTask := &models.ContestTask{
		ContestId: contestId,
		TaskId:    taskId,
	}
	err := tx.Model(&models.ContestTask{}).Create(&contestTask).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) UnAssignTaskFromContest(tx *gorm.DB, contestId, taskId int64) error {
	err := tx.Model(&models.ContestTask{}).Where("contest_id = ? AND task_id = ?", contestId, taskId).Delete(&models.ContestTask{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) IsTaskAssignedToContest(tx *gorm.DB, contestId, taskId int64) (bool, error) {
	var count int64
	err := tx.Model(&models.ContestTask{}).Where("contest_id = ? AND task_id = ?", contestId, taskId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func NewContestRepository(db *gorm.DB) (ContestRepository, error) {
	tables := []interface{}{&models.Contest{}, &models.ContestTask{}}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			err := db.Migrator().CreateTable(table)
			if err != nil {
				return nil, err
			}
		}
	}

	return &contestRepository{}, nil
}
