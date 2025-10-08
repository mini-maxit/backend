package models

// TestCase is a struct that contains the input and output of the test case for task.
type TestCase struct {
	ID           int64 `gorm:"primaryKey"`
	TaskID       int64 `gorm:"not null"`
	InputFileID  int64 `gorm:"not null"`
	OutputFileID int64 `gorm:"not null"`
	Order        int   `gorm:"not null"`
	TimeLimit    int64 `gorm:"not null"`
	MemoryLimit  int64 `gorm:"not null"`
	Task         Task  `gorm:"foreignKey:TaskID; references:ID"`
	InputFile    File  `gorm:"foreignKey:InputFileID; references:ID"`
	OutputFile   File  `gorm:"foreignKey:OutputFileID; references:ID"`
}
