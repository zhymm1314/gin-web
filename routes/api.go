package routes

import (
	"gin-web/app/common/request"
	"gin-web/app/controllers"
	"gin-web/app/middleware"
	"gin-web/app/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetApiGroupRoutes 定义 api 分组路由 (兼容旧代码)
func SetApiGroupRoutes(router *gin.RouterGroup) {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/test", func(c *gin.Context) {
		time.Sleep(5 * time.Second)

		c.String(http.StatusOK, "success")
	})

	router.POST("/user/register", func(c *gin.Context) {
		var form request.Register
		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": request.GetErrorMsg(form, err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	router.POST("/auth/register", controllers.Register)
	router.POST("/auth/login", controllers.Login)

	// Mod相关路由 - 无需认证
	modController := &controllers.ModController{}
	{
		router.GET("/mods/search", modController.Search)         // 搜索mod
		router.GET("/mods/:id", modController.Detail)            // 获取mod详情
		router.GET("/mods/:id/download", modController.Download) // 下载mod
		router.GET("/games", modController.Games)                // 获取游戏列表
		router.GET("/categories", modController.Categories)      // 获取分类列表
	}

	authRouter := router.Group("").Use(middleware.JWTAuth(services.AppGuardName))
	{
		authRouter.POST("/auth/info", controllers.Info)
		authRouter.POST("/auth/logout", controllers.Logout)
	}
}

// SetApiGroupRoutesWithDI 使用依赖注入的路由注册
func SetApiGroupRoutesWithDI(router *gin.RouterGroup, ctrls ...controllers.Controller) {
	// 基础路由
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/test", func(c *gin.Context) {
		time.Sleep(5 * time.Second)
		c.String(http.StatusOK, "success")
	})

	// 自动注册所有控制器
	for _, ctrl := range ctrls {
		controllers.RegisterController(router, ctrl)
	}
}
