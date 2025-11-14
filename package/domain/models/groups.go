package models

type Group struct {
	ID        int64  `gorm:"primaryKey;"`
	Name      string `gorm:"not null;"`
	CreatedBy int64  `gorm:"not null;"`
	Author    User   `gorm:"foreignKey:CreatedBy; references:ID"`
	BaseModel

	Users []User `gorm:"many2many:user_groups;"`
}

type UserGroup struct {
	UserID  int64 `gorm:"primaryKey;"`
	GroupID int64 `gorm:"primaryKey;"`

	User  User  `gorm:"foreignKey:UserID; references:ID"`
	Group Group `gorm:"foreignKey:GroupID; references:ID"`
}
