package models

type Task struct {
	Id        int64  `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"type:varchar(255)"`
	CreatedBy int64  `gorm:"foreignKey:UserID"`
	Author    User   `gorm:"foreignKey:CreatedBy; references:Id"`
}
