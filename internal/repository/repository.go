package repository

import "gin-web/app/models"

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByMobile(mobile string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}
