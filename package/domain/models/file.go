package models

// Example of correct File struct
//
//	File {
//		"ID": 123,
//		"Filename": "example.txt",
//		"Path": "/uploads/example.txt"
//	}
type File struct {
	ID         int64  `gorm:"primaryKey"`
	Filename   string `gorm:"not null"`
	Path       string `gorm:"not null"` // Full path to the file
	Bucket     string `gorm:"not null"` // Bucket where the file is stored
	ServerType string `gorm:"not null"` // Type of the server where the file is stored (e.g., "local", "s3", etc.)
}
