package services

import (
	"gin-web/app/common/request"
	"gin-web/app/common/response"
	"gin-web/app/models"
	"gin-web/global"
	"math"
)

type modService struct{}

var ModService = &modService{}

// SearchMods 搜索mod
func (s *modService) SearchMods(req request.ModSearchRequest) (*response.ModListResponse, error) {
	var mods []models.Mod
	var total int64

	// 构建查询
	db := global.App.DB.Model(&models.Mod{})

	// 预加载关联数据
	db = db.Preload("Game").Preload("Categories")

	// 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ? OR author LIKE ?", keyword, keyword, keyword)
	}

	// 游戏筛选
	if req.GameID > 0 {
		db = db.Where("game_id = ?", req.GameID)
	}

	// 作者筛选
	if req.Author != "" {
		db = db.Where("author LIKE ?", "%"+req.Author+"%")
	}

	// 分类筛选
	if req.CategoryID > 0 {
		db = db.Joins("JOIN gw_mod_categories ON mods.id = gw_mod_categories.mod_id").
			Where("gw_mod_categories.category_id = ?", req.CategoryID)
	}

	// 排序
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	order := req.Order
	if order == "" {
		order = "desc"
	}

	// 验证排序字段
	validSortFields := map[string]bool{
		"rating":         true,
		"download_count": true,
		"created_at":     true,
		"updated_at":     true,
	}
	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	// 验证排序方向
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	db = db.Order(sortBy + " " + order)

	// 获取总数
	db.Count(&total)

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	db = db.Offset(offset).Limit(pageSize)

	// 执行查询
	if err := db.Find(&mods).Error; err != nil {
		return nil, err
	}

	// 转换为响应格式
	modItems := make([]response.ModItem, len(mods))
	for i, mod := range mods {
		// 获取分类名称
		categoryNames := []string{}
		for _, category := range mod.Categories {
			categoryNames = append(categoryNames, category.Name)
		}

		modItems[i] = response.ModItem{
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

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &response.ModListResponse{
		List:       modItems,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetModDetail 获取mod详情
func (s *modService) GetModDetail(id uint) (*response.ModDetailResponse, error) {
	var mod models.Mod

	// 查询mod详情，预加载关联数据
	if err := global.App.DB.Preload("Game").Preload("Categories").First(&mod, id).Error; err != nil {
		return nil, err
	}

	// 增加下载次数
	global.App.DB.Model(&mod).UpdateColumn("download_count", mod.DownloadCount+1)

	// 转换分类为数组格式
	categories := []models.Category{}
	for _, category := range mod.Categories {
		categories = append(categories, category)
	}

	return &response.ModDetailResponse{
		ID:            mod.ID,
		Name:          mod.Name,
		Description:   mod.Description,
		Author:        mod.Author,
		Version:       mod.Version,
		DownloadURL:   mod.DownloadURL,
		Rating:        mod.Rating,
		DownloadCount: mod.DownloadCount + 1, // 返回更新后的值
		FileSize:      mod.FileSize,
		Game:          mod.Game,
		Categories:    categories,
		CreatedAt:     mod.CreatedAt,
		UpdatedAt:     mod.UpdatedAt,
	}, nil
}

// GetGames 获取游戏列表
func (s *modService) GetGames() (*response.GameListResponse, error) {
	var games []models.Game

	if err := global.App.DB.Find(&games).Error; err != nil {
		return nil, err
	}

	return &response.GameListResponse{
		List: games,
	}, nil
}

// GetCategories 获取分类列表
func (s *modService) GetCategories() (*response.CategoryListResponse, error) {
	var categories []models.Category

	if err := global.App.DB.Find(&categories).Error; err != nil {
		return nil, err
	}

	return &response.CategoryListResponse{
		List: categories,
	}, nil
}
