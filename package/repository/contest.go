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
	// CreatePendingRegistration creates a pending registration request
	CreatePendingRegistration(tx *gorm.DB, registration *models.ContestPendingRegistration) (int64, error)
	// IsPendingRegistrationExists checks if pending registration already exists
	IsPendingRegistrationExists(tx *gorm.DB, contestID int64, userID int64) (bool, error)
	// IsUserParticipant checks if user is already a participant
	IsUserParticipant(tx *gorm.DB, contestID int64, userID int64) (bool, error)
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

func (cr *contestRepository) CreatePendingRegistration(tx *gorm.DB, registration *models.ContestPendingRegistration) (int64, error) {
	err := tx.Create(registration).Error
	if err != nil {
		return 0, err
	}
	return registration.ID, nil
}

func (cr *contestRepository) IsPendingRegistrationExists(tx *gorm.DB, contestID int64, userID int64) (bool, error) {
	var count int64
	err := tx.Model(&models.ContestPendingRegistration{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (cr *contestRepository) IsUserParticipant(tx *gorm.DB, contestID int64, userID int64) (bool, error) {
	var userCount int64
	err := tx.Model(&models.ContestParticipant{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Count(&userCount).Error
	if err != nil {
		return false, err
	}
	var groupCount int64
	err = tx.Model(&models.ContestParticipantGroup{}).Where("contest_id = ?", contestID).
		Joins("JOIN user_groups ON contest_participant_groups.group_id = user_groups.group_id").
		Where("user_groups.user_id = ?", userID).
		Count(&groupCount).Error
	if err != nil {
		return false, err
	}
	return userCount > 0 || groupCount > 0, nil
}

func NewContestRepository(db *gorm.DB) (ContestRepository, error) {
	tables := []any{&models.Contest{}, &models.ContestTask{}, &models.ContestParticipant{}, &models.ContestParticipantGroup{}, &models.ContestPendingRegistration{}}
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
