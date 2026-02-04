package services

import (
	"context"
	"errors"
	"gin-web/global"
	"gin-web/utils"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

type jwtService struct {
}

var JwtService = new(jwtService)

// JwtUser 所有需要颁发 token 的用户模型必须实现这个接口
type JwtUser interface {
	GetUid() string
}

// CustomClaims 自定义 Claims
type CustomClaims struct {
	jwt.RegisteredClaims
}

const (
	TokenType    = "bearer"
	AppGuardName = "app"
)

type TokenOutPut struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// CreateToken 生成 Token
func (jwtService *jwtService) CreateToken(guardName string, user JwtUser) (TokenOutPut, *jwt.Token, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		CustomClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.App.Config.Jwt.JwtTtl) * time.Second)),
				ID:        user.GetUid(),
				Issuer:    guardName, // 用于在中间件中区分不同客户端颁发的 token，避免 token 跨端使用
				NotBefore: jwt.NewNumericDate(time.Now().Add(-1000 * time.Second)),
			},
		},
	)

	tokenStr, err := token.SignedString([]byte(global.App.Config.Jwt.Secret))
	if err != nil {
		return TokenOutPut{}, nil, err
	}

	tokenData := TokenOutPut{
		tokenStr,
		int(global.App.Config.Jwt.JwtTtl),
		TokenType,
	}
	return tokenData, token, nil
}

// 获取黑名单缓存 key
func (jwtService *jwtService) getBlackListKey(tokenStr string) string {
	return "jwt_black_list:" + utils.MD5([]byte(tokenStr))
}

// JoinBlackList token 加入黑名单
func (jwtService *jwtService) JoinBlackList(token *jwt.Token) (err error) {
	nowUnix := time.Now().Unix()
	claims := token.Claims.(*CustomClaims)
	expiresAt := claims.ExpiresAt.Time.Unix()
	timer := time.Duration(expiresAt-nowUnix) * time.Second
	// 将 token 剩余时间设置为缓存有效期，并将当前时间作为缓存 value 值
	err = global.App.Redis.SetNX(context.Background(), jwtService.getBlackListKey(token.Raw), nowUnix, timer).Err()
	return
}

// IsInBlacklist token 是否在黑名单中
func (jwtService *jwtService) IsInBlacklist(tokenStr string) bool {
	joinUnixStr, err := global.App.Redis.Get(context.Background(), jwtService.getBlackListKey(tokenStr)).Result()
	joinUnix, err := strconv.ParseInt(joinUnixStr, 10, 64)
	if joinUnixStr == "" || err != nil {
		return false
	}
	// JwtBlacklistGracePeriod 为黑名单宽限时间，避免并发请求失效
	if time.Now().Unix()-joinUnix < global.App.Config.Jwt.JwtBlacklistGracePeriod {
		return false
	}
	return true
}

func (jwtService *jwtService) GetUserInfo(guardName string, id string) (JwtUser, error) {
	switch guardName {
	case AppGuardName:
		return UserServiceLegacy.GetUserInfo(id)
	default:
		return nil, errors.New("guard " + guardName + " does not exist")
	}
}

// ========== 依赖注入版本的 JWT Service ==========

// JwtServiceDI JWT服务 (依赖注入版本)
type JwtServiceDI struct {
	jwtConfig   JwtConfig
	redisClient RedisClient
}

// JwtConfig JWT配置接口
type JwtConfig interface {
	GetSecret() string
	GetTtl() int64
	GetBlacklistGracePeriod() int64
	GetRefreshGracePeriod() int64
}

// RedisClient Redis客户端接口
type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

// NewJwtServiceDI 创建JWT服务实例
func NewJwtServiceDI(jwtConfig JwtConfig, redisClient RedisClient) *JwtServiceDI {
	return &JwtServiceDI{
		jwtConfig:   jwtConfig,
		redisClient: redisClient,
	}
}

// CreateToken 生成 Token
func (s *JwtServiceDI) CreateToken(guardName string, user JwtUser) (TokenOutPut, *jwt.Token, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		CustomClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtConfig.GetTtl()) * time.Second)),
				ID:        user.GetUid(),
				Issuer:    guardName,
				NotBefore: jwt.NewNumericDate(time.Now().Add(-1000 * time.Second)),
			},
		},
	)

	tokenStr, err := token.SignedString([]byte(s.jwtConfig.GetSecret()))
	if err != nil {
		return TokenOutPut{}, nil, err
	}

	tokenData := TokenOutPut{
		tokenStr,
		int(s.jwtConfig.GetTtl()),
		TokenType,
	}
	return tokenData, token, nil
}

// getBlackListKey 获取黑名单缓存 key
func (s *JwtServiceDI) getBlackListKey(tokenStr string) string {
	return "jwt_black_list:" + utils.MD5([]byte(tokenStr))
}

// JoinBlackList token 加入黑名单
func (s *JwtServiceDI) JoinBlackList(token *jwt.Token) (err error) {
	nowUnix := time.Now().Unix()
	claims := token.Claims.(*CustomClaims)
	expiresAt := claims.ExpiresAt.Time.Unix()
	timer := time.Duration(expiresAt-nowUnix) * time.Second
	err = s.redisClient.SetNX(context.Background(), s.getBlackListKey(token.Raw), nowUnix, timer)
	return
}

// IsInBlacklist token 是否在黑名单中
func (s *JwtServiceDI) IsInBlacklist(tokenStr string) bool {
	joinUnixStr, err := s.redisClient.Get(context.Background(), s.getBlackListKey(tokenStr))
	joinUnix, err := strconv.ParseInt(joinUnixStr, 10, 64)
	if joinUnixStr == "" || err != nil {
		return false
	}
	if time.Now().Unix()-joinUnix < s.jwtConfig.GetBlacklistGracePeriod() {
		return false
	}
	return true
}
