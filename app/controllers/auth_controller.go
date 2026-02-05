package controllers

import (
	"gin-web/app/dto"
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
// @Summary      用户注册
// @Description  创建新用户账号
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "注册信息"
// @Success      200 {object} dto.Response "成功"
// @Failure      400 {object} dto.Response "参数错误"
// @Router       /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var form dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&form); err != nil {
		dto.ValidateFail(ctx, dto.GetErrorMsg(form, err))
		return
	}

	user, err := c.userService.Register(form)
	if err != nil {
		dto.BusinessFail(ctx, err.Error())
		return
	}
	dto.Success(ctx, user)
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户登录获取 JWT Token
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "登录信息"
// @Success      200 {object} dto.Response "成功返回 Token"
// @Failure      400 {object} dto.Response "参数错误"
// @Failure      401 {object} dto.Response "认证失败"
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var form dto.LoginRequest
	if err := ctx.ShouldBindJSON(&form); err != nil {
		dto.ValidateFail(ctx, dto.GetErrorMsg(form, err))
		return
	}

	user, err := c.userService.Login(form)
	if err != nil {
		dto.BusinessFail(ctx, err.Error())
		return
	}

	tokenData, _, err := c.jwtService.CreateToken(services.AppGuardName, user)
	if err != nil {
		dto.BusinessFail(ctx, err.Error())
		return
	}
	dto.Success(ctx, tokenData)
}

// Info 获取用户信息
// @Summary      获取用户信息
// @Description  获取当前登录用户的详细信息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} dto.Response "成功"
// @Failure      401 {object} dto.Response "未授权"
// @Router       /auth/info [post]
func (c *AuthController) Info(ctx *gin.Context) {
	user, err := c.userService.GetUserInfo(ctx.Keys["id"].(string))
	if err != nil {
		dto.BusinessFail(ctx, err.Error())
		return
	}
	dto.Success(ctx, user)
}

// Logout 用户登出
// @Summary      用户登出
// @Description  用户登出，将当前 Token 加入黑名单
// @Tags         认证
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} dto.Response "成功"
// @Failure      401 {object} dto.Response "未授权"
// @Router       /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	err := c.jwtService.JoinBlackList(ctx.Keys["token"].(*jwt.Token))
	if err != nil {
		dto.BusinessFail(ctx, "登出失败")
		return
	}
	dto.Success(ctx, nil)
}
