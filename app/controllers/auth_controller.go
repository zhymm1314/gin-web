package controllers

import (
	"gin-web/app/common/request"
	"gin-web/app/common/response"
	"gin-web/app/middleware"
	"gin-web/app/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthController 认证控制器
type AuthController struct {
	userService   *services.UserService
	jwtService    *services.JwtService
	jwtMiddleware *middleware.JwtMiddleware
}

// NewAuthController 创建认证控制器实例
func NewAuthController(userService *services.UserService, jwtService *services.JwtService, jwtMiddleware *middleware.JwtMiddleware) *AuthController {
	return &AuthController{
		userService:   userService,
		jwtService:    jwtService,
		jwtMiddleware: jwtMiddleware,
	}
}

// Prefix 返回路由前缀
func (c *AuthController) Prefix() string {
	return "/auth"
}

// Routes 返回路由列表
func (c *AuthController) Routes() []Route {
	return []Route{
		{Method: "POST", Path: "/register", Handler: c.Register},
		{Method: "POST", Path: "/login", Handler: c.Login},
		{Method: "POST", Path: "/info", Handler: c.Info, Middlewares: []gin.HandlerFunc{c.jwtMiddleware.JWTAuth(services.AppGuardName)}},
		{Method: "POST", Path: "/logout", Handler: c.Logout, Middlewares: []gin.HandlerFunc{c.jwtMiddleware.JWTAuth(services.AppGuardName)}},
	}
}

// Register 用户注册
func (c *AuthController) Register(ctx *gin.Context) {
	var form request.Register
	if err := ctx.ShouldBindJSON(&form); err != nil {
		response.ValidateFail(ctx, request.GetErrorMsg(form, err))
		return
	}

	user, err := c.userService.Register(form)
	if err != nil {
		response.BusinessFail(ctx, err.Error())
		return
	}
	response.Success(ctx, user)
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var form request.Login
	if err := ctx.ShouldBindJSON(&form); err != nil {
		response.ValidateFail(ctx, request.GetErrorMsg(form, err))
		return
	}

	user, err := c.userService.Login(form)
	if err != nil {
		response.BusinessFail(ctx, err.Error())
		return
	}

	tokenData, _, err := c.jwtService.CreateToken(services.AppGuardName, user)
	if err != nil {
		response.BusinessFail(ctx, err.Error())
		return
	}
	response.Success(ctx, tokenData)
}

// Info 获取用户信息
func (c *AuthController) Info(ctx *gin.Context) {
	user, err := c.userService.GetUserInfo(ctx.Keys["id"].(string))
	if err != nil {
		response.BusinessFail(ctx, err.Error())
		return
	}
	response.Success(ctx, user)
}

// Logout 用户登出
func (c *AuthController) Logout(ctx *gin.Context) {
	err := c.jwtService.JoinBlackList(ctx.Keys["token"].(*jwt.Token))
	if err != nil {
		response.BusinessFail(ctx, "登出失败")
		return
	}
	response.Success(ctx, nil)
}
