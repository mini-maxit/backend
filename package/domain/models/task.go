package models

type Task struct {
	ID                int64  `gorm:"primaryKey;autoIncrement"`
	Title             string `gorm:"type:varchar(255)"`
	DescriptionFileID int64  `gorm:"null;default:null"`
	CreatedBy         int64  `gorm:"foreignKey:UserID"`
	IsVisible         bool   `gorm:"default:false;not null"`

	BaseModel

	Author          User `gorm:"foreignKey:CreatedBy; references:ID"`
	DescriptionFile File `gorm:"foreignKey:DescriptionFileID; references:ID"`
}
