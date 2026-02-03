package controllers

import "github.com/gin-gonic/gin"

// Route 路由定义
type Route struct {
	Method      string
	Path        string
	Handler     gin.HandlerFunc
	Middlewares []gin.HandlerFunc
}

// Controller 控制器接口
type Controller interface {
	// Prefix 返回路由前缀
	Prefix() string
	// Routes 返回路由列表
	Routes() []Route
}

// RegisterController 注册控制器路由
func RegisterController(router *gin.RouterGroup, controller Controller) {
	group := router.Group(controller.Prefix())
	for _, route := range controller.Routes() {
		handlers := append(route.Middlewares, route.Handler)
		switch route.Method {
		case "GET":
			group.GET(route.Path, handlers...)
		case "POST":
			group.POST(route.Path, handlers...)
		case "PUT":
			group.PUT(route.Path, handlers...)
		case "DELETE":
			group.DELETE(route.Path, handlers...)
		case "PATCH":
			group.PATCH(route.Path, handlers...)
		}
	}
}
