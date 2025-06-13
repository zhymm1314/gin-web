package models

import (
	"time"
)

// Mod mod模型
type Mod struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	Name          string  `json:"name" gorm:"size:255;not null;index" binding:"required"`
	Description   string  `json:"description" gorm:"type:text"`
	Author        string  `json:"author" gorm:"size:100;index"`
	Version       string  `json:"version" gorm:"size:50"`
	DownloadURL   string  `json:"download_url" gorm:"size:500"`
	ImageURL      string  `json:"image_url" gorm:"size:500"`
	Rating        float64 `json:"rating" gorm:"type:decimal(3,2);default:0"`
	DownloadCount int     `json:"download_count" gorm:"default:0;index"`
	FileSize      int64   `json:"file_size" gorm:"default:0"`

	// 外键关联
	GameID     uint       `json:"game_id" gorm:"not null;index"`
	Game       Game       `json:"game" gorm:"foreignKey:GameID"`
	Categories []Category `json:"categories" gorm:"many2many:gw_mod_categories;"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Mod) TableName() string {
	return "mods"
}
