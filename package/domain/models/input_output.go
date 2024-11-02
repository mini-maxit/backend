package models

// InputOutput is a struct that contains the input and output of the task
type InputOutput struct {
	Id          uint    `gorm:"primaryKey"`
	TaskId      uint    `gorm:"not null"`
	Order       int     `gorm:"not null"`
	TimeLimit   float64 `gorm:"not null"`
	MemoryLimit float64 `gorm:"not null"`
	Task        Task    `gorm:"foreignKey:TaskId; references:Id"`
}
