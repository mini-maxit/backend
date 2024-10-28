package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type UserSolution interface {
	CreateUserSolution(tx *gorm.DB, userSolution models.UserSolution) (uint, error)
	MarkUserSolutionProcessing(tx *gorm.DB, userSolutionID int) error
	MarkUserSolutionComplete(tx *gorm.DB, userSolutionID uint) error
	MarkUserSolutionFailed(db *gorm.DB, userSolutionID uint, errorMsg error) error
}

type UserSolutionRepository struct{}

func (us *UserSolutionRepository) CreateUserSolution(tx *gorm.DB, userSolution models.UserSolution) (uint, error) {
	err := tx.Create(userSolution).Error
	if err != nil {
		return 0, err
	}
	return userSolution.ID, nil
}

func (us *UserSolutionRepository) MarkUserSolutionProcessing(tx *gorm.DB, userSolutionID int) error {
	err := tx.Model(&models.UserSolution{}).Where("id = ?", userSolutionID).Update("status", "processing").Error
	return err
}

func (us *UserSolutionRepository) MarkUserSolutionComplete(tx *gorm.DB, userSolutionID uint) error {
	err := tx.Model(&models.UserSolution{}).Where("id = ?", userSolutionID).Update("status", "completed").Error
	return err
}

func (us *UserSolutionRepository) MarkUserSolutionFailed(db *gorm.DB, userSolutionID uint, errorMsg error) error {
	stringErrorMsg := errorMsg.Error()
	err := db.Model(&models.UserSolution{}).Where("id = ?", userSolutionID).Updates(map[string]interface{}{
		"status":         "failed",
		"status_message": stringErrorMsg,
	}).Error
	return err
}

func NewUserSolutionRepository(db *gorm.DB) (UserSolution, error) {
	if !db.Migrator().HasTable(&models.UserSolution{}) {
		err := db.Migrator().CreateTable(&models.UserSolution{})
		if err != nil {
			return nil, err
		}
	}
	return &UserSolutionRepository{}, nil
}
