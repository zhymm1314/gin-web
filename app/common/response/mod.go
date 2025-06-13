package response

import (
	"gin-web/app/models"
	"time"
)

// ModListResponse mod列表响应
type ModListResponse struct {
	List       []ModItem `json:"list"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

// ModItem mod列表项
type ModItem struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Author        string    `json:"author"`
	Version       string    `json:"version"`
	Rating        float64   `json:"rating"`
	DownloadCount int       `json:"download_count"`
	FileSize      int64     `json:"file_size"`
	GameName      string    `json:"game_name"`
	Categories    []string  `json:"categories"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ModDetailResponse mod详情响应
type ModDetailResponse struct {
	ID            uint              `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Author        string            `json:"author"`
	Version       string            `json:"version"`
	DownloadURL   string            `json:"download_url"`
	Rating        float64           `json:"rating"`
	DownloadCount int               `json:"download_count"`
	FileSize      int64             `json:"file_size"`
	Game          models.Game       `json:"game"`
	Categories    []models.Category `json:"categories"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// GameListResponse 游戏列表响应
type GameListResponse struct {
	List []models.Game `json:"list"`
}

// CategoryListResponse 分类列表响应
type CategoryListResponse struct {
	List []models.Category `json:"list"`
}
