package dto

import (
	"gin-web/app/models"
	"time"
)

// ================================
// Mod 模块 DTO（Data Transfer Object）
// ================================

// -------------------- Request --------------------

// ModSearchRequest 搜索 Mod 请求
// @Description Mod 搜索筛选条件
type ModSearchRequest struct {
	Keyword    string `form:"keyword" json:"keyword" example:"武器"`                                                      // 搜索关键词
	GameID     uint   `form:"game_id" json:"game_id" example:"1"`                                                       // 游戏ID
	CategoryID uint   `form:"category_id" json:"category_id" example:"2"`                                               // 分类ID
	Author     string `form:"author" json:"author" example:"ModAuthor"`                                                 // 作者名称
	SortBy     string `form:"sort_by" json:"sort_by" example:"download_count" enums:"rating,download_count,created_at"` // 排序字段
	Order      string `form:"order" json:"order" example:"desc" enums:"asc,desc"`                                       // 排序方向
	Page       int    `form:"page" json:"page" binding:"min=0" example:"1"`                                             // 页码
	PageSize   int    `form:"page_size" json:"page_size" binding:"min=0,max=100" example:"20"`                          // 每页数量
}

// GetMessages 自定义验证错误信息
func (r ModSearchRequest) GetMessages() ValidatorMessages {
	return ValidatorMessages{
		"Page.min":     "页码不能小于0",
		"PageSize.min": "每页数量不能小于0",
		"PageSize.max": "每页数量不能超过100",
	}
}

// ModDetailRequest 获取 Mod 详情请求
// @Description 通过 ID 获取 Mod 详情
type ModDetailRequest struct {
	ID uint `uri:"id" binding:"required,min=1" example:"1"` // Mod ID
}

// GetMessages 自定义验证错误信息
func (r ModDetailRequest) GetMessages() ValidatorMessages {
	return ValidatorMessages{
		"ID.required": "Mod ID 不能为空",
		"ID.min":      "Mod ID 必须大于0",
	}
}

// -------------------- Response --------------------

// ModListResponse Mod 列表响应
// @Description Mod 分页列表
type ModListResponse struct {
	List       []ModItemResponse `json:"list"`        // Mod 列表
	Total      int64             `json:"total"`       // 总数
	Page       int               `json:"page"`        // 当前页
	PageSize   int               `json:"page_size"`   // 每页数量
	TotalPages int               `json:"total_pages"` // 总页数
}

// ModItemResponse Mod 列表项
// @Description Mod 基本信息（列表展示用）
type ModItemResponse struct {
	ID            uint      `json:"id" example:"1"`                 // Mod ID
	Name          string    `json:"name" example:"超级武器包"`           // Mod 名称
	Author        string    `json:"author" example:"ModAuthor"`     // 作者
	Version       string    `json:"version" example:"1.0.0"`        // 版本号
	Rating        float64   `json:"rating" example:"4.5"`           // 评分
	DownloadCount int       `json:"download_count" example:"10000"` // 下载次数
	FileSize      int64     `json:"file_size" example:"1048576"`    // 文件大小（字节）
	GameName      string    `json:"game_name" example:"GTA5"`       // 游戏名称
	Categories    []string  `json:"categories" example:"武器,载具"`     // 分类列表
	CreatedAt     time.Time `json:"created_at"`                     // 创建时间
	UpdatedAt     time.Time `json:"updated_at"`                     // 更新时间
}

// ModDetailResponse Mod 详情响应
// @Description Mod 完整详情信息
type ModDetailResponse struct {
	ID            uint              `json:"id" example:"1"`                 // Mod ID
	Name          string            `json:"name" example:"超级武器包"`           // Mod 名称
	Description   string            `json:"description" example:"这是一个..."`  // 详细描述
	Author        string            `json:"author" example:"ModAuthor"`     // 作者
	Version       string            `json:"version" example:"1.0.0"`        // 版本号
	DownloadURL   string            `json:"download_url"`                   // 下载链接
	Rating        float64           `json:"rating" example:"4.5"`           // 评分
	DownloadCount int               `json:"download_count" example:"10000"` // 下载次数
	FileSize      int64             `json:"file_size" example:"1048576"`    // 文件大小（字节）
	Game          models.Game       `json:"game"`                           // 所属游戏
	Categories    []models.Category `json:"categories"`                     // 分类列表
	CreatedAt     time.Time         `json:"created_at"`                     // 创建时间
	UpdatedAt     time.Time         `json:"updated_at"`                     // 更新时间
}

// GameListResponse 游戏列表响应
// @Description 所有支持的游戏列表
type GameListResponse struct {
	List []models.Game `json:"list"` // 游戏列表
}

// CategoryListResponse 分类列表响应
// @Description 所有 Mod 分类列表
type CategoryListResponse struct {
	List []models.Category `json:"list"` // 分类列表
}
