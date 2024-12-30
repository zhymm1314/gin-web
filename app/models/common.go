package models

import (
	"gorm.io/gorm"
	"time"
)

// 自增ID主键
type ID struct {
	ID uint `json:"id" gorm:"primaryKey"`
}

// Timestamps 创建、更新时间
type Timestamps struct {
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// 软删除
type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
