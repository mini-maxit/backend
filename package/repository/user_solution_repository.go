package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type UserSolutionRepository interface {
	GetUserSolution(tx *gorm.DB, userSolutionID int64) (*models.UserSolution, error)
	CreateUserSolution(tx *gorm.DB, userSolution models.UserSolution) (int64, error)
	CreateUserSolutionResult(tx *gorm.DB, userSolutionResult models.UserSolutionResult) (int64, error)
	MarkUserSolutionProcessing(tx *gorm.DB, userSolutionID int64) error
	MarkUserSolutionComplete(tx *gorm.DB, userSolutionID int64) error
	MarkUserSolutionFailed(db *gorm.DB, userSolutionID int64, errorMsg string) error
}

type UserSolutionRepositoryImpl struct{}

func (us *UserSolutionRepositoryImpl) GetUserSolution(tx *gorm.DB, userSolutionID int64) (*models.UserSolution, error) {
	var userSolution models.UserSolution
	err := tx.Where("id = ?", userSolutionID).First(&userSolution).Error
	if err != nil {
		return nil, err
	}
	return &userSolution, nil
}

func (us *UserSolutionRepositoryImpl) CreateUserSolution(tx *gorm.DB, userSolution models.UserSolution) (int64, error) {
	err := tx.Create(&userSolution).Error
	if err != nil {
		return 0, err
	}
	return userSolution.Id, nil
}

func (us *UserSolutionRepositoryImpl) CreateUserSolutionResult(tx *gorm.DB, userSolutionResult models.UserSolutionResult) (int64, error) {
	err := tx.Create(&userSolutionResult).Error
	if err != nil {
		return -1, err
	}
	return userSolutionResult.Id, nil
}

func (us *UserSolutionRepositoryImpl) MarkUserSolutionProcessing(tx *gorm.DB, userSolutionID int64) error {
	err := tx.Model(&models.UserSolution{}).Where("id = ?", userSolutionID).Update("status", "processing").Error
	return err
}

func (us *UserSolutionRepositoryImpl) MarkUserSolutionComplete(tx *gorm.DB, userSolutionID int64) error {
	err := tx.Model(&models.UserSolution{}).Where("id = ?", userSolutionID).Update("status", "completed").Error
	return err
}

func (us *UserSolutionRepositoryImpl) MarkUserSolutionFailed(db *gorm.DB, userSolutionID int64, errorMsg string) error {
	err := db.Model(&models.UserSolution{}).Where("id = ?", userSolutionID).Updates(map[string]interface{}{
		"status":         "failed",
		"status_message": errorMsg,
	}).Error
	return err
}

func NewUserSolutionRepository(db *gorm.DB) (UserSolutionRepository, error) {
	if !db.Migrator().HasTable(&models.UserSolution{}) {
		err := db.Migrator().CreateTable(&models.UserSolution{})
		if err != nil {
			return nil, err
		}
	}
	return &UserSolutionRepositoryImpl{}, nil
}
