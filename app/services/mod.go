package services

import (
	"go.uber.org/zap"

	"gin-web/app/dto"
	"gin-web/app/models"
	"gin-web/internal/repository"
)

// ModService Mod服务
type ModService struct {
	repo repository.ModRepository
	log  *zap.Logger
}

// NewModService 创建Mod服务实例
func NewModService(repo repository.ModRepository, log *zap.Logger) *ModService {
	return &ModService{repo: repo, log: log}
}

// SearchMods 搜索mod
func (s *ModService) SearchMods(req dto.ModSearchRequest) (*dto.ModListResponse, error) {
	// 转换 DTO 为 Repository 查询条件
	criteria := repository.ModSearchCriteria{
		Keyword:    req.Keyword,
		GameID:     req.GameID,
		CategoryID: req.CategoryID,
		Author:     req.Author,
		SortBy:     req.SortBy,
		Order:      req.Order,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// 调用 Repository 执行搜索
	result, err := s.repo.Search(criteria)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	modItems := make([]dto.ModItemResponse, len(result.Mods))
	for i, mod := range result.Mods {
		categoryNames := make([]string, len(mod.Categories))
		for j, category := range mod.Categories {
			categoryNames[j] = category.Name
		}

		modItems[i] = dto.ModItemResponse{
			ID:            mod.ID,
			Name:          mod.Name,
			Author:        mod.Author,
			Version:       mod.Version,
			Rating:        mod.Rating,
			DownloadCount: mod.DownloadCount,
			FileSize:      mod.FileSize,
			GameName:      mod.Game.Name,
			Categories:    categoryNames,
			CreatedAt:     mod.CreatedAt,
			UpdatedAt:     mod.UpdatedAt,
		}
	}

	return &dto.ModListResponse{
		List:       modItems,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

// GetModDetail 获取mod详情
func (s *ModService) GetModDetail(id uint) (*dto.ModDetailResponse, error) {
	mod, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 增加下载次数
	_ = s.repo.UpdateDownloadCount(mod)

	// 转换分类为数组格式
	categories := make([]models.Category, len(mod.Categories))
	copy(categories, mod.Categories)

	return &dto.ModDetailResponse{
		ID:            mod.ID,
		Name:          mod.Name,
		Description:   mod.Description,
		Author:        mod.Author,
		Version:       mod.Version,
		DownloadURL:   mod.DownloadURL,
		Rating:        mod.Rating,
		DownloadCount: mod.DownloadCount + 1,
		FileSize:      mod.FileSize,
		Game:          mod.Game,
		Categories:    categories,
		CreatedAt:     mod.CreatedAt,
		UpdatedAt:     mod.UpdatedAt,
	}, nil
}

// GetGames 获取游戏列表
func (s *ModService) GetGames() (*dto.GameListResponse, error) {
	games, err := s.repo.FindAllGames()
	if err != nil {
		return nil, err
	}

	return &dto.GameListResponse{
		List: games,
	}, nil
}

// GetCategories 获取分类列表
func (s *ModService) GetCategories() (*dto.CategoryListResponse, error) {
	categories, err := s.repo.FindAllCategories()
	if err != nil {
		return nil, err
	}

	return &dto.CategoryListResponse{
		List: categories,
	}, nil
}
