package models

type Task struct {
	ID                int64  `gorm:"primaryKey;autoIncrement"`
	Title             string `gorm:"type:varchar(255)"`
	DescriptionFileID int64  `gorm:"null;default:null"`
	CreatedBy         int64  `gorm:"foreignKey:UserID"`

	baseModel

	Author          User    `gorm:"foreignKey:CreatedBy; references:ID"`
	Groups          []Group `gorm:"many2many:task_groups;"`
	DescriptionFile File    `gorm:"foreignKey:DescriptionFileID; references:ID"`
}

type TaskUser struct {
	TaskID int64 `gorm:"primaryKey"`
	UserID int64 `gorm:"primaryKey"`
}
