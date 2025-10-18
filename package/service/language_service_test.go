package service_test

import (
	"testing"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

var trueValue = true

var falseValue = false

func TestLanguageServiceInit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)
	ls := service.NewLanguageService(lr)
	tx := &gorm.DB{}

	workerLanguages := schemas.HandShakeResponsePayload{
		Languages: []struct {
			Name      string   `json:"name"`
			Versions  []string `json:"versions"`
			Extension string   `json:"extension"`
		}{
			{
				Name:      "python",
				Versions:  []string{"3.9", "3.10"},
				Extension: ".py",
			},
			{
				Name:      "javascript",
				Versions:  []string{"18", "20"},
				Extension: ".js",
			},
		},
	}

	t.Run("Success with new languages", func(t *testing.T) {
		// Mock that no existing languages are found
		lr.EXPECT().GetAll(tx).Return([]models.LanguageConfig{}, nil).Times(1)

		// Expect creates for each language-version combination
		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "python",
			Version:       "3.9",
			FileExtension: ".py",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "python",
			Version:       "3.10",
			FileExtension: ".py",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "javascript",
			Version:       "18",
			FileExtension: ".js",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "javascript",
			Version:       "20",
			FileExtension: ".js",
		}).Return(nil).Times(1)

		err := ls.Init(tx, workerLanguages)
		require.NoError(t, err)
	})

	t.Run("Success with existing enabled languages", func(t *testing.T) {
		existingLanguages := []models.LanguageConfig{
			{ID: 1, Type: "python", Version: "3.9", FileExtension: ".py", IsDisabled: &falseValue},
			{ID: 2, Type: "python", Version: "3.10", FileExtension: ".py", IsDisabled: &trueValue},
		}

		lr.EXPECT().GetAll(tx).Return(existingLanguages, nil).Times(1)

		// Expect enabling of disabled language
		lr.EXPECT().MarkEnabled(tx, int64(2)).Return(nil).Times(1)
		lr.EXPECT().MarkEnabled(tx, int64(1)).Return(nil).Times(1)

		// Expect creates for new language-version combinations
		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "javascript",
			Version:       "18",
			FileExtension: ".js",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "javascript",
			Version:       "20",
			FileExtension: ".js",
		}).Return(nil).Times(1)

		err := ls.Init(tx, workerLanguages)
		require.NoError(t, err)
	})

	t.Run("Success with languages to disable", func(t *testing.T) {
		existingLanguages := []models.LanguageConfig{
			{ID: 1, Type: "python", Version: "3.9", FileExtension: ".py", IsDisabled: &falseValue},
			{ID: 2, Type: "go", Version: "1.19", FileExtension: ".go", IsDisabled: &falseValue}, // This should be disabled
		}

		lr.EXPECT().GetAll(tx).Return(existingLanguages, nil).Times(1)

		lr.EXPECT().MarkEnabled(tx, int64(1)).Return(nil).Times(1)

		// Expect creates for new language-version combinations
		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "python",
			Version:       "3.10",
			FileExtension: ".py",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "javascript",
			Version:       "18",
			FileExtension: ".js",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "javascript",
			Version:       "20",
			FileExtension: ".js",
		}).Return(nil).Times(1)

		// Expect disabling of language not in worker languages
		lr.EXPECT().MarkDisabled(tx, int64(2)).Return(nil).Times(1)

		err := ls.Init(tx, workerLanguages)
		require.NoError(t, err)
	})

	t.Run("Error getting existing languages", func(t *testing.T) {
		lr.EXPECT().GetAll(tx).Return(nil, assert.AnError).Times(1)

		err := ls.Init(tx, workerLanguages)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Error creating new language", func(t *testing.T) {
		lr.EXPECT().GetAll(tx).Return([]models.LanguageConfig{}, nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "python",
			Version:       "3.9",
			FileExtension: ".py",
		}).Return(assert.AnError).Times(1)

		err := ls.Init(tx, workerLanguages)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Error marking language disabled", func(t *testing.T) {
		existingLanguages := []models.LanguageConfig{
			{ID: 1, Type: "go", Version: "1.19", FileExtension: ".go", IsDisabled: &falseValue}, // This should be disabled
		}

		lr.EXPECT().GetAll(tx).Return(existingLanguages, nil).Times(1)

		// Expect creates for new language-version combinations
		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "python",
			Version:       "3.9",
			FileExtension: ".py",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "python",
			Version:       "3.10",
			FileExtension: ".py",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "javascript",
			Version:       "18",
			FileExtension: ".js",
		}).Return(nil).Times(1)

		lr.EXPECT().Create(tx, &models.LanguageConfig{
			Type:          "javascript",
			Version:       "20",
			FileExtension: ".js",
		}).Return(nil).Times(1)

		// Error when marking language as disabled
		lr.EXPECT().MarkDisabled(tx, int64(1)).Return(assert.AnError).Times(1)

		err := ls.Init(tx, workerLanguages)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestLanguageServiceGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)
	ls := service.NewLanguageService(lr)
	tx := &gorm.DB{}

	t.Run("Success with languages", func(t *testing.T) {
		languages := []models.LanguageConfig{
			{ID: 1, Type: "python", Version: "3.9", FileExtension: ".py", IsDisabled: &falseValue},
			{ID: 2, Type: "javascript", Version: "18", FileExtension: ".js", IsDisabled: &falseValue},
		}

		lr.EXPECT().GetAll(tx).Return(languages, nil).Times(1)

		result, err := ls.GetAll(tx)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "python", result[0].Type)
		assert.Equal(t, "3.9", result[0].Version)
		assert.Equal(t, ".py", result[0].FileExtension)
		assert.Equal(t, int64(2), result[1].ID)
		assert.Equal(t, "javascript", result[1].Type)
		assert.Equal(t, "18", result[1].Version)
		assert.Equal(t, ".js", result[1].FileExtension)
	})

	t.Run("Success with no languages", func(t *testing.T) {
		lr.EXPECT().GetAll(tx).Return([]models.LanguageConfig{}, nil).Times(1)

		result, err := ls.GetAll(tx)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error getting languages", func(t *testing.T) {
		lr.EXPECT().GetAll(tx).Return(nil, assert.AnError).Times(1)

		result, err := ls.GetAll(tx)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestLanguageServiceGetAllEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)
	ls := service.NewLanguageService(lr)
	tx := &gorm.DB{}

	t.Run("Success with enabled languages", func(t *testing.T) {
		languages := []models.LanguageConfig{
			{ID: 1, Type: "python", Version: "3.9", FileExtension: ".py", IsDisabled: &falseValue},
			{ID: 3, Type: "java", Version: "17", FileExtension: ".java", IsDisabled: &falseValue},
		}

		lr.EXPECT().GetEnabled(tx).Return(languages, nil).Times(1)

		result, err := ls.GetAllEnabled(tx)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "python", result[0].Type)
		assert.Equal(t, "3.9", result[0].Version)
		assert.Equal(t, ".py", result[0].FileExtension)
		assert.Equal(t, int64(3), result[1].ID)
		assert.Equal(t, "java", result[1].Type)
		assert.Equal(t, "17", result[1].Version)
		assert.Equal(t, ".java", result[1].FileExtension)
	})

	t.Run("Success with no enabled languages", func(t *testing.T) {
		lr.EXPECT().GetEnabled(tx).Return([]models.LanguageConfig{}, nil).Times(1)

		result, err := ls.GetAllEnabled(tx)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error getting enabled languages", func(t *testing.T) {
		lr.EXPECT().GetEnabled(tx).Return(nil, assert.AnError).Times(1)

		result, err := ls.GetAllEnabled(tx)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestLanguageToSchema(t *testing.T) {
	t.Run("Convert model to schema", func(t *testing.T) {
		language := &models.LanguageConfig{
			ID:            1,
			Type:          "python",
			Version:       "3.9",
			FileExtension: ".py",
			IsDisabled:    &falseValue,
		}

		result := service.LanguageToSchema(language)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, "python", result.Type)
		assert.Equal(t, "3.9", result.Version)
		assert.Equal(t, ".py", result.FileExtension)
	})
}

func TestNewLanguageService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)

	t.Run("Create new language service", func(t *testing.T) {
		ls := service.NewLanguageService(lr)
		assert.NotNil(t, ls)
	})
}
