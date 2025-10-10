package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type File interface {
	Create(tx *gorm.DB, file *models.File) error
}

type file struct {
}

func (f *file) Create(tx *gorm.DB, file *models.File) error {
	err := tx.Model(models.File{}).Create(&file).Error
	if err != nil {
		return err
	}
	return nil
}

func NewFileRepository(db *gorm.DB) (File, error) {
	tables := []any{&models.File{}}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			err := db.Migrator().CreateTable(table)
			if err != nil {
				return nil, err
			}
		}
	}

	return &file{}, nil
}
