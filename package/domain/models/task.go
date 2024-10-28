package models

type Task struct {
	ID            int64  `gorm:"primaryKey;autoIncrement"`
	Title         string `gorm:"type:varchar(255)"`
	DirPath       string `gorm:"type:varchar(255)"`
	InputDirPath  string `gorm:"type:varchar(255)"`
	OutputDirPath string `gorm:"type:varchar(255)"`
}
