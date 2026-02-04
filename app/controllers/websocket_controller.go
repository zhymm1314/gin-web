package controllers

import (
	"gin-web/pkg/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebSocketController WebSocket 控制器
type WebSocketController struct {
	manager *websocket.Manager
}

// NewWebSocketController 创建 WebSocket 控制器实例
func NewWebSocketController(manager *websocket.Manager) *WebSocketController {
	return &WebSocketController{manager: manager}
}

// Prefix 返回路由前缀
func (c *WebSocketController) Prefix() string {
	return "/ws"
}

// Routes 返回路由列表
func (c *WebSocketController) Routes() []Route {
	return []Route{
		{Method: "GET", Path: "/connect", Handler: c.Connect},
		{Method: "GET", Path: "/status", Handler: c.Status},
		{Method: "POST", Path: "/broadcast", Handler: c.Broadcast},
		{Method: "POST", Path: "/send", Handler: c.SendToUser},
	}
}

// Connect WebSocket 连接
// @Summary      WebSocket 连接
// @Description  建立 WebSocket 连接
// @Tags         WebSocket
// @Param        user_id query string false "用户ID"
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws/connect [get]
func (c *WebSocketController) Connect(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	// 或从 JWT 中间件获取: userID := ctx.GetString("id")

	if err := c.manager.HandleRequest(ctx.Writer, ctx.Request, userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// Status 获取 WebSocket 状态
// @Summary      WebSocket 状态
// @Description  获取在线人数等状态信息
// @Tags         WebSocket
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /ws/status [get]
func (c *WebSocketController) Status(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"online_count": c.manager.OnlineCount(),
		"online_users": c.manager.OnlineUsers(),
	})
}

// Broadcast 广播消息
// @Summary      广播消息
// @Description  向所有在线用户广播消息
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        message body websocket.Message true "消息内容"
// @Success      200 {object} map[string]interface{}
// @Router       /ws/broadcast [post]
func (c *WebSocketController) Broadcast(ctx *gin.Context) {
	var msg websocket.Message
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.manager.Broadcast(&msg)
	ctx.JSON(http.StatusOK, gin.H{"message": "broadcast sent"})
}

// SendToUser 发送消息给指定用户
// @Summary      发送消息给用户
// @Description  向指定用户发送消息
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        message body websocket.Message true "消息内容"
// @Success      200 {object} map[string]interface{}
// @Router       /ws/send [post]
func (c *WebSocketController) SendToUser(ctx *gin.Context) {
	var msg websocket.Message
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg.To == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "to field is required"})
		return
	}

	c.manager.SendToUser(msg.To, &msg)
	ctx.JSON(http.StatusOK, gin.H{"message": "message sent"})
}
