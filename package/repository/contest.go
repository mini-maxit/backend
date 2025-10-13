package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type Contest interface {
}

type contest struct{}

func NewContestRepository(db *gorm.DB) (Contest, error) {
	tables := []any{&models.Contest{}, &models.ContestTask{}, &models.ContestParticipant{}}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			err := db.Migrator().CreateTable(table)
			if err != nil {
				return nil, err
			}
		}
	}
	return &contest{}, nil
}
