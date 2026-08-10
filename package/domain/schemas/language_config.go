package schemas

type LanguageConfig struct {
	ID            int64  `json:"id"`
	Type          string `json:"language"`
	Version       string `json:"version"`
	FileExtension string `json:"fileExtension"`
	IsDisabled    bool   `json:"isDisabled"`
}
