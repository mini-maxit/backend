package models

type User struct {
	Id           int64  `gorm:"primaryKey;autoIncrement"`
	Name         string `gorm:"NOT NULL"`
	Surname      string `gorm:"NOT NULL"`
	Email        string `gorm:"NOT NULL;UNIQUE"`
	Username     string `gorm:"NOT NULL;UNIQUE"`
	PasswordHash string `gorm:"NOT NULL"`
	Role         string
}

type UserRole string

const (
	UserRoleStudent UserRole = "student"
	UserRoleTeacher UserRole = "teacher"
	UserRoleAdmin   UserRole = "admin"
)
