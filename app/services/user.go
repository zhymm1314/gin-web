package services

import (
	"gin-web/app/common/request"
	"gin-web/app/models"
	"gin-web/global"
	"gin-web/internal/repository"
	bizErr "gin-web/pkg/errors"
	"gin-web/utils"
	"go.uber.org/zap"
	"strconv"
)

// UserService 用户服务 (依赖注入版本)
type UserService struct {
	repo repository.UserRepository
	log  *zap.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(repo repository.UserRepository, log *zap.Logger) *UserService {
	return &UserService{repo: repo, log: log}
}

// Register 注册
func (s *UserService) Register(params request.Register) (*models.User, error) {
	existUser, _ := s.repo.FindByMobile(params.Mobile)
	if existUser != nil {
		return nil, bizErr.ErrUserExists
	}

	user := &models.User{
		Name:     params.Name,
		Mobile:   params.Mobile,
		Password: utils.BcryptMake([]byte(params.Password)),
	}

	if err := s.repo.Create(user); err != nil {
		s.log.Error("create user failed", zap.Error(err))
		return nil, bizErr.Wrap(err, bizErr.CodeInternalError, "创建用户失败")
	}

	return user, nil
}

// Login 登录
func (s *UserService) Login(params request.Login) (*models.User, error) {
	user, err := s.repo.FindByMobile(params.Mobile)
	if err != nil {
		return nil, bizErr.ErrUserNotFound
	}
	if !utils.BcryptMakeCheck([]byte(params.Password), user.Password) {
		return nil, bizErr.ErrPassword
	}
	return user, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(id string) (*models.User, error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		return nil, bizErr.Wrap(err, bizErr.CodeValidationError, "无效的用户ID")
	}
	user, err := s.repo.FindByID(uint(intId))
	if err != nil {
		return nil, bizErr.ErrUserNotFound
	}
	return user, nil
}

// ========== 兼容旧代码的全局变量方式 ==========

type userServiceLegacy struct {
}

var UserServiceLegacy = new(userServiceLegacy)

// Register 注册 (兼容旧代码)
func (userService *userServiceLegacy) Register(params request.Register) (err error, user models.User) {
	var result = global.App.DB.Where("mobile = ?", params.Mobile).Select("id").First(&models.User{})
	if result.RowsAffected != 0 {
		err = bizErr.ErrUserExists
		return
	}
	user = models.User{Name: params.Name, Mobile: params.Mobile, Password: utils.BcryptMake([]byte(params.Password))}
	err = global.App.DB.Create(&user).Error
	return
}

// Login 登录 (兼容旧代码)
func (userService *userServiceLegacy) Login(params request.Login) (err error, user *models.User) {
	err = global.App.DB.Where("mobile = ?", params.Mobile).First(&user).Error
	if err != nil || !utils.BcryptMakeCheck([]byte(params.Password), user.Password) {
		err = bizErr.New(bizErr.CodePasswordError, "用户名不存在或密码错误")
	}
	return
}

// GetUserInfo 获取用户信息 (兼容旧代码)
func (userService *userServiceLegacy) GetUserInfo(id string) (err error, user models.User) {
	intId, err := strconv.Atoi(id)
	err = global.App.DB.First(&user, intId).Error
	if err != nil {
		err = bizErr.ErrUserNotFound
	}
	return
}
