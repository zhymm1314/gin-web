package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"gin-web/app/dto"
	"gin-web/app/models"
	"gin-web/app/services"
	"gin-web/internal/repository"
)

// MockModRepository Mod 仓储 Mock
type MockModRepository struct {
	mock.Mock
}

func (m *MockModRepository) Search(criteria repository.ModSearchCriteria) (*repository.ModSearchResult, error) {
	args := m.Called(criteria)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ModSearchResult), args.Error(1)
}

func (m *MockModRepository) FindByID(id uint) (*models.Mod, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Mod), args.Error(1)
}

func (m *MockModRepository) UpdateDownloadCount(mod *models.Mod) error {
	args := m.Called(mod)
	return args.Error(0)
}

func (m *MockModRepository) FindAllGames() ([]models.Game, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Game), args.Error(1)
}

func (m *MockModRepository) FindAllCategories() ([]models.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func TestModService_SearchMods_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockModRepository)
	logger, _ := zap.NewDevelopment()
	service := services.NewModService(mockRepo, logger)

	req := dto.ModSearchRequest{
		Keyword:  "test",
		Page:     1,
		PageSize: 10,
	}

	now := time.Now()
	expectedResult := &repository.ModSearchResult{
		Mods: []models.Mod{
			{
				Name:          "Test Mod 1",
				Author:        "Author1",
				Version:       "1.0.0",
				Rating:        4.5,
				DownloadCount: 100,
				FileSize:      1024,
				Game:          models.Game{Name: "Game1"},
				Categories:    []models.Category{{Name: "Category1"}},
			},
		},
		Total:      1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}
	expectedResult.Mods[0].ID = 1
	expectedResult.Mods[0].CreatedAt = now
	expectedResult.Mods[0].UpdatedAt = now

	criteria := repository.ModSearchCriteria{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	mockRepo.On("Search", criteria).Return(expectedResult, nil)

	// Act
	result, err := service.SearchMods(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.List, 1)
	assert.Equal(t, "Test Mod 1", result.List[0].Name)
	assert.Equal(t, "Game1", result.List[0].GameName)
	assert.Equal(t, []string{"Category1"}, result.List[0].Categories)
	assert.Equal(t, int64(1), result.Total)
	mockRepo.AssertExpectations(t)
}

func TestModService_SearchMods_Empty(t *testing.T) {
	// Arrange
	mockRepo := new(MockModRepository)
	logger, _ := zap.NewDevelopment()
	service := services.NewModService(mockRepo, logger)

	req := dto.ModSearchRequest{
		Keyword:  "nonexistent",
		Page:     1,
		PageSize: 10,
	}

	expectedResult := &repository.ModSearchResult{
		Mods:       []models.Mod{},
		Total:      0,
		Page:       1,
		PageSize:   10,
		TotalPages: 0,
	}

	criteria := repository.ModSearchCriteria{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	mockRepo.On("Search", criteria).Return(expectedResult, nil)

	// Act
	result, err := service.SearchMods(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.List, 0)
	assert.Equal(t, int64(0), result.Total)
	mockRepo.AssertExpectations(t)
}

func TestModService_SearchMods_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockModRepository)
	logger, _ := zap.NewDevelopment()
	service := services.NewModService(mockRepo, logger)

	req := dto.ModSearchRequest{
		Page:     1,
		PageSize: 10,
	}

	criteria := repository.ModSearchCriteria{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	mockRepo.On("Search", criteria).Return(nil, errors.New("database error"))

	// Act
	result, err := service.SearchMods(req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestModService_GetModDetail_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockModRepository)
	logger, _ := zap.NewDevelopment()
	service := services.NewModService(mockRepo, logger)

	now := time.Now()
	mod := &models.Mod{
		Name:          "Test Mod",
		Description:   "Description",
		Author:        "Author",
		Version:       "1.0.0",
		DownloadURL:   "https://example.com/download",
		Rating:        4.5,
		DownloadCount: 100,
		FileSize:      1024,
		Game:          models.Game{Name: "Game1"},
		Categories:    []models.Category{{Name: "Category1"}},
	}
	mod.ID = 1
	mod.CreatedAt = now
	mod.UpdatedAt = now

	mockRepo.On("FindByID", uint(1)).Return(mod, nil)
	mockRepo.On("UpdateDownloadCount", mod).Return(nil)

	// Act
	result, err := service.GetModDetail(1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Mod", result.Name)
	assert.Equal(t, 101, result.DownloadCount) // +1
	mockRepo.AssertExpectations(t)
}

func TestModService_GetModDetail_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockModRepository)
	logger, _ := zap.NewDevelopment()
	service := services.NewModService(mockRepo, logger)

	mockRepo.On("FindByID", uint(999)).Return(nil, errors.New("not found"))

	// Act
	result, err := service.GetModDetail(999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestModService_GetGames_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockModRepository)
	logger, _ := zap.NewDevelopment()
	service := services.NewModService(mockRepo, logger)

	games := []models.Game{
		{Name: "Game1"},
		{Name: "Game2"},
	}

	mockRepo.On("FindAllGames").Return(games, nil)

	// Act
	result, err := service.GetGames()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.List, 2)
	mockRepo.AssertExpectations(t)
}

func TestModService_GetCategories_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockModRepository)
	logger, _ := zap.NewDevelopment()
	service := services.NewModService(mockRepo, logger)

	categories := []models.Category{
		{Name: "Category1"},
		{Name: "Category2"},
	}

	mockRepo.On("FindAllCategories").Return(categories, nil)

	// Act
	result, err := service.GetCategories()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.List, 2)
	mockRepo.AssertExpectations(t)
}
