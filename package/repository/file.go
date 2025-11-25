package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type File interface {
	Create(db database.Database, file *models.File) error
}

type file struct {
}

func (f *file) Create(db database.Database, file *models.File) error {
	tx := db.GetInstance()
	err := tx.Model(models.File{}).Create(&file).Error
	if err != nil {
		return err
	}
	return nil
}

func NewFileRepository() File {
	return &file{}
}
