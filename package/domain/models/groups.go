package models

type Group struct {
	ID        int64  `gorm:"primaryKey;"`
	Name      string `gorm:"not null;"`
	CreatedBy int64  `gorm:"not null;"`
	Author    User   `gorm:"foreignKey:CreatedBy; references:ID"`
	BaseModel

	Tasks []Task `gorm:"many2many:task_groups;"`
	Users []User `gorm:"many2many:user_groups;"`
}

type UserGroup struct {
	UserID  int64 `gorm:"primaryKey;"`
	GroupID int64 `gorm:"primaryKey;"`

	User  User  `gorm:"foreignKey:UserID; references:ID"`
	Group Group `gorm:"foreignKey:GroupID; references:ID"`
}

type TaskGroup struct {
	TaskID  int64 `gorm:"primaryKey;"`
	GroupID int64 `gorm:"primaryKey;"`

	Task  Task  `gorm:"foreignKey:TaskID; references:ID"`
	Group Group `gorm:"foreignKey:GroupID; references:ID"`
}
