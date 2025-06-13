package models

import (
	"time"
)

// Game 游戏模型
type Game struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:255;not null;index" binding:"required"`
	EnglishName string    `json:"english_name" gorm:"size:255;index"`
	Description string    `json:"description" gorm:"type:text"`
	CoverImage  string    `json:"cover_image" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Game) TableName() string {
	return "games"
}
