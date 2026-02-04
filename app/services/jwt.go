package services

import (
	"context"
	"errors"
	"gin-web/utils"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

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

// TokenOutPut Token 输出结构
type TokenOutPut struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
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

// UserGetter 用户获取接口
type UserGetter interface {
	GetUserInfo(id string) (JwtUser, error)
}

// JwtService JWT服务
type JwtService struct {
	jwtConfig   JwtConfig
	redisClient RedisClient
	userGetter  UserGetter
}

// NewJwtService 创建JWT服务实例
func NewJwtService(jwtConfig JwtConfig, redisClient RedisClient, userGetter UserGetter) *JwtService {
	return &JwtService{
		jwtConfig:   jwtConfig,
		redisClient: redisClient,
		userGetter:  userGetter,
	}
}

// CreateToken 生成 Token
func (s *JwtService) CreateToken(guardName string, user JwtUser) (TokenOutPut, *jwt.Token, error) {
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
func (s *JwtService) getBlackListKey(tokenStr string) string {
	return "jwt_black_list:" + utils.MD5([]byte(tokenStr))
}

// JoinBlackList token 加入黑名单
func (s *JwtService) JoinBlackList(token *jwt.Token) error {
	nowUnix := time.Now().Unix()
	claims := token.Claims.(*CustomClaims)
	expiresAt := claims.ExpiresAt.Time.Unix()
	timer := time.Duration(expiresAt-nowUnix) * time.Second
	return s.redisClient.SetNX(context.Background(), s.getBlackListKey(token.Raw), nowUnix, timer)
}

// IsInBlacklist token 是否在黑名单中
func (s *JwtService) IsInBlacklist(tokenStr string) bool {
	joinUnixStr, err := s.redisClient.Get(context.Background(), s.getBlackListKey(tokenStr))
	if err != nil || joinUnixStr == "" {
		return false
	}
	joinUnix, err := strconv.ParseInt(joinUnixStr, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-joinUnix < s.jwtConfig.GetBlacklistGracePeriod() {
		return false
	}
	return true
}

// GetUserInfo 获取用户信息
func (s *JwtService) GetUserInfo(guardName string, id string) (JwtUser, error) {
	if s.userGetter == nil {
		return nil, errors.New("user getter not configured")
	}
	switch guardName {
	case AppGuardName:
		return s.userGetter.GetUserInfo(id)
	default:
		return nil, errors.New("guard " + guardName + " does not exist")
	}
}

// GetSecret 获取 JWT 密钥
func (s *JwtService) GetSecret() string {
	return s.jwtConfig.GetSecret()
}

// GetRefreshGracePeriod 获取刷新宽限期
func (s *JwtService) GetRefreshGracePeriod() int64 {
	return s.jwtConfig.GetRefreshGracePeriod()
}

// GetBlacklistGracePeriod 获取黑名单宽限期
func (s *JwtService) GetBlacklistGracePeriod() int64 {
	return s.jwtConfig.GetBlacklistGracePeriod()
}
