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

func NewFileRepository() File {
	return &file{}
}
