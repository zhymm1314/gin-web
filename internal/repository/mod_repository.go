package repository

import (
	"gin-web/app/models"
	"gorm.io/gorm"
)

// ModSearchCriteria 搜索条件（封装查询参数，避免 Service 直接操作 gorm.DB）
type ModSearchCriteria struct {
	Keyword    string
	GameID     uint
	CategoryID uint
	Author     string
	SortBy     string // rating, download_count, created_at, updated_at
	Order      string // asc, desc
	Page       int
	PageSize   int
}

// ModSearchResult 搜索结果
type ModSearchResult struct {
	Mods       []models.Mod
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// ModRepository Mod仓储接口
type ModRepository interface {
	Search(criteria ModSearchCriteria) (*ModSearchResult, error)
	FindByID(id uint) (*models.Mod, error)
	UpdateDownloadCount(mod *models.Mod) error
	FindAllGames() ([]models.Game, error)
	FindAllCategories() ([]models.Category, error)
}

type modRepository struct {
	db *gorm.DB
}

// NewModRepository 创建 Mod 仓储实例
func NewModRepository(db *gorm.DB) ModRepository {
	return &modRepository{db: db}
}

// Search 搜索 Mod（查询构建逻辑封装在 Repository 内部）
func (r *modRepository) Search(criteria ModSearchCriteria) (*ModSearchResult, error) {
	db := r.db.Model(&models.Mod{})

	// 预加载关联数据
	db = db.Preload("Game").Preload("Categories")

	// 关键词搜索
	if criteria.Keyword != "" {
		keyword := "%" + criteria.Keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ? OR author LIKE ?", keyword, keyword, keyword)
	}

	// 游戏筛选
	if criteria.GameID > 0 {
		db = db.Where("game_id = ?", criteria.GameID)
	}

	// 作者筛选
	if criteria.Author != "" {
		db = db.Where("author LIKE ?", "%"+criteria.Author+"%")
	}

	// 分类筛选
	if criteria.CategoryID > 0 {
		db = db.Joins("JOIN gw_mod_categories ON mods.id = gw_mod_categories.mod_id").
			Where("gw_mod_categories.category_id = ?", criteria.CategoryID)
	}

	// 排序
	sortBy := criteria.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	order := criteria.Order
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
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页
	page := criteria.Page
	if page < 1 {
		page = 1
	}
	pageSize := criteria.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	db = db.Offset(offset).Limit(pageSize)

	// 执行查询
	var mods []models.Mod
	if err := db.Find(&mods).Error; err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &ModSearchResult{
		Mods:       mods,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *modRepository) FindByID(id uint) (*models.Mod, error) {
	var mod models.Mod
	if err := r.db.Preload("Game").Preload("Categories").First(&mod, id).Error; err != nil {
		return nil, err
	}
	return &mod, nil
}

func (r *modRepository) UpdateDownloadCount(mod *models.Mod) error {
	return r.db.Model(mod).UpdateColumn("download_count", mod.DownloadCount+1).Error
}

func (r *modRepository) FindAllGames() ([]models.Game, error) {
	var games []models.Game
	if err := r.db.Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
}

func (r *modRepository) FindAllCategories() ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
