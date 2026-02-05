package services

import (
	"strconv"

	"go.uber.org/zap"

	"gin-web/app/dto"
	"gin-web/app/models"
	"gin-web/internal/repository"
	bizErr "gin-web/pkg/errors"
	"gin-web/utils"
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
func (s *UserService) Register(params dto.RegisterRequest) (*models.User, error) {
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
func (s *UserService) Login(params dto.LoginRequest) (*models.User, error) {
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
