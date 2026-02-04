package models

import (
	"gorm.io/gorm"
	"time"
)

// ID 主键
type ID struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
}

// Timestamps 时间戳
type Timestamps struct {
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SoftDeletes 软删除
type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BaseModel 基础模型（组合使用）
type BaseModel struct {
	ID
	Timestamps
	SoftDeletes
}
