package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"gin-web/app/dto"
	"gin-web/app/services"
)

// JwtMiddleware JWT中间件依赖
type JwtMiddleware struct {
	jwtService *services.JwtService
}

// NewJwtMiddleware 创建JWT中间件实例
func NewJwtMiddleware(jwtService *services.JwtService) *JwtMiddleware {
	return &JwtMiddleware{jwtService: jwtService}
}

// JWTAuth 创建JWT认证中间件
func (m *JwtMiddleware) JWTAuth(guardName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Request.Header.Get("Authorization")
		if tokenStr == "" {
			dto.TokenFail(c)
			c.Abort()
			return
		}
		tokenStr = tokenStr[len(services.TokenType)+1:]

		// Token 解析校验
		token, err := jwt.ParseWithClaims(tokenStr, &services.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtService.GetSecret()), nil
		})
		if err != nil || !token.Valid || m.jwtService.IsInBlacklist(tokenStr) {
			dto.TokenFail(c)
			c.Abort()
			return
		}

		claims := token.Claims.(*services.CustomClaims)
		// Token 发布者校验
		if claims.Issuer != guardName {
			dto.TokenFail(c)
			c.Abort()
			return
		}

		// token 续签
		if claims.ExpiresAt.Time.Unix()-time.Now().Unix() < m.jwtService.GetRefreshGracePeriod() {
			user, err := m.jwtService.GetUserInfo(guardName, claims.ID)
			if err == nil {
				tokenData, _, _ := m.jwtService.CreateToken(guardName, user)
				c.Header("new-token", tokenData.AccessToken)
				c.Header("new-expires-in", strconv.Itoa(tokenData.ExpiresIn))
				_ = m.jwtService.JoinBlackList(token)
			}
		}

		c.Set("token", token)
		c.Set("id", claims.ID)
	}
}
