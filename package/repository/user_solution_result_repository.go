package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type UserSolutionResult interface {
	CreateUserSolutionResult(tx *gorm.DB, solutionResult models.UserSolutionResult) (int64, error)
}

type UserSolutionResultRepository struct{}

// Store the result of the solution in the database
func (usr *UserSolutionResultRepository) CreateUserSolutionResult(tx *gorm.DB, solutionResult models.UserSolutionResult) (int64, error) {
	if err := tx.Create(&solutionResult).Error; err != nil {
		return 0, err
	}
	return solutionResult.Id, nil
}

func NewUserSolutionResultRepository(db *gorm.DB) (UserSolutionResult, error) {
	if !db.Migrator().HasTable(&models.UserSolutionResult{}) {
		if err := db.Migrator().CreateTable(&models.UserSolutionResult{}); err != nil {
			return nil, err
		}
	}
	return &UserSolutionResultRepository{}, nil

}
