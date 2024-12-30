package models

import (
	"strconv"
)

type User struct {
	ID
	Name     string `json:"name1" gorm:"not null;comment:用户名称"`
	Mobile   string `json:"mobile2" gorm:"not null;index;comment:用户手机号"`
	Password string `json:"-" gorm:"not null;default:'';comment:用户密码"` //密码不反回
	Timestamps
	SoftDeletes
}

func (user User) GetUid() string {
	return strconv.Itoa(int(user.ID.ID))
}
