package models

// InputOutput is a struct that contains the input and output of the task
type InputOutput struct {
	ID          uint    `gorm:"primaryKey"`
	TaskID      uint    `gorm:"not null: foreignKey:TaskID"`
	Order       int     `gorm:"not null"`
	TimeLimit   float64 `gorm:"not null"`
	MemoryLimit float64 `gorm:"not null"`
}
