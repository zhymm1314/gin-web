package routes

import (
	"gin-web/app/controllers"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetApiGroupRoutes 定义 api 分组路由
func SetApiGroupRoutes(router *gin.RouterGroup, ctrls ...controllers.Controller) {
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
