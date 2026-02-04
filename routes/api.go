package routes

import (
	"gin-web/app/common/request"
	"gin-web/app/controllers"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetApiGroupRoutes 定义 api 分组路由 (Legacy 版本，仅保留基础路由)
// 注意: 此函数已弃用，请使用 SetApiGroupRoutesWithDI
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
