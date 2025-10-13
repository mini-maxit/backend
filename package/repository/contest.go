package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type ContestRepository interface {
	// Create creates a new contest
	Create(tx *gorm.DB, contest *models.Contest) (int64, error)
	// Get retrieves a contest by ID
	Get(tx *gorm.DB, contestID int64) (*models.Contest, error)
	// GetAll retrieves all contests with pagination and sorting
	GetAll(tx *gorm.DB, offset int, limit int, sort string) ([]models.Contest, error)
	// GetAllForCreator retrieves all contests created by a specific user with pagination and sorting
	GetAllForCreator(tx *gorm.DB, creatorID int64, offset int, limit int, sort string) ([]models.Contest, error)
	// Edit updates a contest
	Edit(tx *gorm.DB, contestID int64, contest *models.Contest) (*models.Contest, error)
	// Delete removes a contest
	Delete(tx *gorm.DB, contestID int64) error
}

type contestRepository struct{}

func (cr *contestRepository) Create(tx *gorm.DB, contest *models.Contest) (int64, error) {
	err := tx.Create(contest).Error
	if err != nil {
		return 0, err
	}
	return contest.ID, nil
}

func (cr *contestRepository) Get(tx *gorm.DB, contestID int64) (*models.Contest, error) {
	var contest models.Contest
	err := tx.Where("id = ?", contestID).First(&contest).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (cr *contestRepository) GetAll(tx *gorm.DB, offset int, limit int, sort string) ([]models.Contest, error) {
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

func (cr *contestRepository) GetAllForCreator(tx *gorm.DB, creatorID int64, offset int, limit int, sort string) ([]models.Contest, error) {
	var contests []models.Contest
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}
	err = tx.Model(&models.Contest{}).Where("created_by = ?", creatorID).Find(&contests).Error
	if err != nil {
		return nil, err
	}
	return contests, nil
}

func (cr *contestRepository) Edit(tx *gorm.DB, contestID int64, contest *models.Contest) (*models.Contest, error) {
	err := tx.Model(&models.Contest{}).Where("id = ?", contestID).Updates(contest).Error
	if err != nil {
		return nil, err
	}
	return cr.Get(tx, contestID)
}

func (cr *contestRepository) Delete(tx *gorm.DB, contestID int64) error {
	err := tx.Where("id = ?", contestID).Delete(&models.Contest{}).Error
	if err != nil {
		return err
	}
	return nil
}

func NewContestRepository(db *gorm.DB) (ContestRepository, error) {
	tables := []any{&models.Contest{}, &models.ContestTask{}, &models.ContestParticipant{}, &models.ContestParticipantGroup{}}
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
