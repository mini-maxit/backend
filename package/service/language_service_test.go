package service_test

import (
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	testPythonName       = "python"
	testPythonVersion310 = "3.10"
	testPythonVersion39  = "3.9"
	testPythonExtension  = ".py"
	testJSName           = "javascript"
	testJSExtension      = ".js"
	testCPPName          = "CPP"
	testCPPExtension     = "cpp"
)

var trueValue = true

var falseValue = false

func TestLanguageServiceInit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)
	ls := service.NewLanguageService(lr)
	db := &testutils.MockDatabase{}

	workerLanguages := schemas.HandShakeResponsePayload{
		Languages: []struct {
			Name      string   `json:"name"`
			Versions  []string `json:"versions"`
			Extension string   `json:"extension"`
		}{
			{
				Name:      testPythonName,
				Versions:  []string{testPythonVersion39, testPythonVersion310},
				Extension: testPythonExtension,
			},
			{
				Name:      testJSName,
				Versions:  []string{"18", "20"},
				Extension: testJSExtension,
			},
		},
	}

	t.Run("Success with new languages", func(t *testing.T) {
		// Mock that no existing languages are found
		lr.EXPECT().GetAll(db).Return([]models.LanguageConfig{}, nil).Times(1)

		// Expect creates for each language-version combination
		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testPythonName,
			Version:       testPythonVersion39,
			FileExtension: testPythonExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testPythonName,
			Version:       testPythonVersion310,
			FileExtension: testPythonExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testJSName,
			Version:       "18",
			FileExtension: testJSExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testJSName,
			Version:       "20",
			FileExtension: testJSExtension,
		}).Return(nil).Times(1)

		err := ls.Init(db, workerLanguages)
		require.NoError(t, err)
	})

	t.Run("Success with existing enabled languages", func(t *testing.T) {
		existingLanguages := []models.LanguageConfig{
			{ID: 1, Type: testPythonName, Version: testPythonVersion39, FileExtension: testPythonExtension, IsDisabled: &falseValue},
			{ID: 2, Type: testPythonName, Version: testPythonVersion310, FileExtension: testPythonExtension, IsDisabled: &trueValue},
		}

		lr.EXPECT().GetAll(db).Return(existingLanguages, nil).Times(1)

		// Expect creates for new language-version combinations
		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testJSName,
			Version:       "18",
			FileExtension: testJSExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testJSName,
			Version:       "20",
			FileExtension: testJSExtension,
		}).Return(nil).Times(1)

		err := ls.Init(db, workerLanguages)
		require.NoError(t, err)
	})

	t.Run("Success with languages to disable", func(t *testing.T) {
		existingLanguages := []models.LanguageConfig{
			{ID: 1, Type: testPythonName, Version: testPythonVersion39, FileExtension: testPythonExtension, IsDisabled: &falseValue},
			{ID: 2, Type: "go", Version: "1.19", FileExtension: ".go", IsDisabled: &falseValue}, // This should be disabled
		}

		lr.EXPECT().GetAll(db).Return(existingLanguages, nil).Times(1)

		// Expect creates for new language-version combinations
		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testPythonName,
			Version:       testPythonVersion310,
			FileExtension: testPythonExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testJSName,
			Version:       "18",
			FileExtension: testJSExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testJSName,
			Version:       "20",
			FileExtension: testJSExtension,
		}).Return(nil).Times(1)

		// Expect disabling of language not in worker languages
		lr.EXPECT().MarkDisabled(db, int64(2)).Return(nil).Times(1)

		err := ls.Init(db, workerLanguages)
		require.NoError(t, err)
	})

	t.Run("Error getting existing languages", func(t *testing.T) {
		lr.EXPECT().GetAll(db).Return(nil, assert.AnError).Times(1)

		err := ls.Init(db, workerLanguages)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Error creating new language", func(t *testing.T) {
		lr.EXPECT().GetAll(db).Return([]models.LanguageConfig{}, nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testPythonName,
			Version:       testPythonVersion39,
			FileExtension: testPythonExtension,
		}).Return(assert.AnError).Times(1)

		err := ls.Init(db, workerLanguages)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Error marking language disabled", func(t *testing.T) {
		existingLanguages := []models.LanguageConfig{
			{ID: 1, Type: "go", Version: "1.19", FileExtension: ".go", IsDisabled: &falseValue}, // This should be disabled
		}

		lr.EXPECT().GetAll(db).Return(existingLanguages, nil).Times(1)

		// Expect creates for new language-version combinations
		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testPythonName,
			Version:       testPythonVersion39,
			FileExtension: testPythonExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testPythonName,
			Version:       testPythonVersion310,
			FileExtension: testPythonExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testJSName,
			Version:       "18",
			FileExtension: testJSExtension,
		}).Return(nil).Times(1)

		lr.EXPECT().Create(db, &models.LanguageConfig{
			Type:          testJSName,
			Version:       "20",
			FileExtension: testJSExtension,
		}).Return(nil).Times(1)

		// Error when marking language as disabled
		lr.EXPECT().MarkDisabled(db, int64(1)).Return(assert.AnError).Times(1)

		err := ls.Init(db, workerLanguages)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestLanguageServiceGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)
	ls := service.NewLanguageService(lr)
	db := &testutils.MockDatabase{}

	t.Run("Success with languages", func(t *testing.T) {
		languages := []models.LanguageConfig{
			{ID: 1, Type: testPythonName, Version: testPythonVersion39, FileExtension: testPythonExtension, IsDisabled: &falseValue},
			{ID: 2, Type: testJSName, Version: "18", FileExtension: testJSExtension, IsDisabled: &falseValue},
		}

		lr.EXPECT().GetAll(db).Return(languages, nil).Times(1)

		result, err := ls.GetAll(db)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		// Sorted by type asc: javascript before python.
		assert.Equal(t, int64(2), result[0].ID)
		assert.Equal(t, testJSName, result[0].Type)
		assert.Equal(t, "18", result[0].Version)
		assert.Equal(t, testJSExtension, result[0].FileExtension)
		assert.Equal(t, int64(1), result[1].ID)
		assert.Equal(t, testPythonName, result[1].Type)
		assert.Equal(t, testPythonVersion39, result[1].Version)
		assert.Equal(t, testPythonExtension, result[1].FileExtension)
	})

	t.Run("Success with no languages", func(t *testing.T) {
		lr.EXPECT().GetAll(db).Return([]models.LanguageConfig{}, nil).Times(1)

		result, err := ls.GetAll(db)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error getting languages", func(t *testing.T) {
		lr.EXPECT().GetAll(db).Return(nil, assert.AnError).Times(1)

		result, err := ls.GetAll(db)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestLanguageServiceGetAll_DeterministicOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)
	ls := service.NewLanguageService(lr)
	db := &testutils.MockDatabase{}

	// Repository returns unsorted rows; the service must order by type then version.
	languages := []models.LanguageConfig{
		{ID: 1, Type: testCPPName, Version: "20", FileExtension: testCPPExtension, IsDisabled: &trueValue},
		{ID: 2, Type: testPythonName, Version: testPythonVersion310, FileExtension: testPythonExtension, IsDisabled: &falseValue},
		{ID: 3, Type: testCPPName, Version: "11", FileExtension: testCPPExtension, IsDisabled: &falseValue},
		{ID: 4, Type: testPythonName, Version: testPythonVersion310, FileExtension: testPythonExtension, IsDisabled: &falseValue},
	}
	lr.EXPECT().GetAll(db).Return(languages, nil).Times(1)

	result, err := ls.GetAll(db)
	require.NoError(t, err)
	require.Len(t, result, 4)

	// Sorted by type (asc), then version (asc).
	assert.Equal(t, int64(3), result[0].ID, "CPP 11 first")
	assert.Equal(t, int64(1), result[1].ID, "CPP 20 second")
	assert.Equal(t, int64(2), result[2].ID, "python 3.10 third")
	assert.Equal(t, int64(4), result[3].ID, "python 3.12 last")
}

func TestLanguageServiceGetAllEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)
	ls := service.NewLanguageService(lr)
	db := &testutils.MockDatabase{}

	t.Run("Success with enabled languages", func(t *testing.T) {
		languages := []models.LanguageConfig{
			{ID: 1, Type: testPythonName, Version: testPythonVersion39, FileExtension: testPythonExtension, IsDisabled: &falseValue},
			{ID: 3, Type: "java", Version: "17", FileExtension: ".java", IsDisabled: &falseValue},
		}

		lr.EXPECT().GetEnabled(db).Return(languages, nil).Times(1)

		result, err := ls.GetAllEnabled(db)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, testPythonName, result[0].Type)
		assert.Equal(t, testPythonVersion39, result[0].Version)
		assert.Equal(t, testPythonExtension, result[0].FileExtension)
		assert.Equal(t, int64(3), result[1].ID)
		assert.Equal(t, "java", result[1].Type)
		assert.Equal(t, "17", result[1].Version)
		assert.Equal(t, ".java", result[1].FileExtension)
	})

	t.Run("Success with no enabled languages", func(t *testing.T) {
		lr.EXPECT().GetEnabled(db).Return([]models.LanguageConfig{}, nil).Times(1)

		result, err := ls.GetAllEnabled(db)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error getting enabled languages", func(t *testing.T) {
		lr.EXPECT().GetEnabled(db).Return(nil, assert.AnError).Times(1)

		result, err := ls.GetAllEnabled(db)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestLanguageToSchema(t *testing.T) {
	t.Run("Convert model to schema", func(t *testing.T) {
		language := &models.LanguageConfig{
			ID:            1,
			Type:          testPythonName,
			Version:       testPythonVersion39,
			FileExtension: testPythonExtension,
			IsDisabled:    &falseValue,
		}

		result := service.LanguageToSchema(language)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, testPythonName, result.Type)
		assert.Equal(t, testPythonVersion39, result.Version)
		assert.Equal(t, testPythonExtension, result.FileExtension)
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

func TestLanguageServiceToggleLanguageVisibility(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lr := mock_repository.NewMockLanguageRepository(ctrl)
	ls := service.NewLanguageService(lr)
	db := &testutils.MockDatabase{}

	disabled := true
	enabled := false

	t.Run("Toggle disabled language to enabled", func(t *testing.T) {
		languages := []models.LanguageConfig{
			{ID: 1, Type: testPythonName, Version: testPythonVersion39, FileExtension: testPythonExtension, IsDisabled: &disabled},
		}
		lr.EXPECT().GetAll(db).Return(languages, nil).Times(1)
		lr.EXPECT().MarkEnabled(db, int64(1)).Return(nil).Times(1)

		err := ls.ToggleLanguageVisibility(db, 1)
		require.NoError(t, err)
	})

	t.Run("Toggle enabled language to disabled", func(t *testing.T) {
		languages := []models.LanguageConfig{
			{ID: 2, Type: testJSName, Version: "18", FileExtension: testJSExtension, IsDisabled: &enabled},
		}
		lr.EXPECT().GetAll(db).Return(languages, nil).Times(1)
		lr.EXPECT().MarkDisabled(db, int64(2)).Return(nil).Times(1)

		err := ls.ToggleLanguageVisibility(db, 2)
		require.NoError(t, err)
	})

	t.Run("Language not found", func(t *testing.T) {
		lr.EXPECT().GetAll(db).Return([]models.LanguageConfig{}, nil).Times(1)

		err := ls.ToggleLanguageVisibility(db, 99)
		require.Error(t, err)
	})

	t.Run("GetAll error", func(t *testing.T) {
		lr.EXPECT().GetAll(db).Return(nil, assert.AnError).Times(1)

		err := ls.ToggleLanguageVisibility(db, 1)
		require.Error(t, err)
	})
}

func TestLanguageToSchema_IsDisabled(t *testing.T) {
	disabled := true

	t.Run("Disabled language mapped", func(t *testing.T) {
		language := &models.LanguageConfig{
			ID: 1, Type: testPythonName, Version: testPythonVersion39,
			FileExtension: testPythonExtension, IsDisabled: &disabled,
		}
		result := service.LanguageToSchema(language)
		assert.True(t, result.IsDisabled)
	})

	t.Run("Nil IsDisabled defaults to false", func(t *testing.T) {
		language := &models.LanguageConfig{
			ID: 2, Type: testJSName, Version: "18", FileExtension: testJSExtension,
		}
		result := service.LanguageToSchema(language)
		assert.False(t, result.IsDisabled)
	})
}
