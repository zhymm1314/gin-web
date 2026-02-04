package models

import (
	"strconv"
)

// User 用户模型
type User struct {
	ID
	Name     string `json:"name" gorm:"type:varchar(100);not null;comment:用户名称"`
	Mobile   string `json:"mobile" gorm:"type:varchar(20);not null;uniqueIndex;comment:用户手机号"`
	Password string `json:"-" gorm:"type:varchar(255);not null;comment:用户密码"`
	Timestamps
	SoftDeletes
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// GetUid 获取用户ID字符串（实现 JwtUser 接口）
func (u User) GetUid() string {
	return strconv.FormatUint(uint64(u.ID.ID), 10)
}

// MaskMobile 获取脱敏手机号
func (u User) MaskMobile() string {
	if len(u.Mobile) < 11 {
		return u.Mobile
	}
	return u.Mobile[:3] + "****" + u.Mobile[7:]
}
