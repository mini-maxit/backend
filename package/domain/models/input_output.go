package models

// InputOutput is a struct that contains the input and output of the task.
type InputOutput struct {
	ID          int64 `gorm:"primaryKey"`
	TaskID      int64 `gorm:"not null"`
	Order       int   `gorm:"not null"`
	TimeLimit   int64 `gorm:"not null"`
	MemoryLimit int64 `gorm:"not null"`
	Task        Task  `gorm:"foreignKey:TaskID; references:ID"`
}
