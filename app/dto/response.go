package dto

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"gin-web/config"
	"gin-web/global"
)

// ================================
// 统一响应结构
// ================================

// Response 统一响应结构体
// @Description API 统一响应格式
type Response struct {
	ErrorCode int         `json:"error_code" example:"0"` // 错误码，0 表示成功
	Data      interface{} `json:"data"`                   // 响应数据
	Message   string      `json:"message" example:"ok"`   // 响应消息
}

// PaginationResponse 分页响应
// @Description 通用分页结构
type PaginationResponse struct {
	List       interface{} `json:"list"`        // 数据列表
	Total      int64       `json:"total"`       // 总数
	Page       int         `json:"page"`        // 当前页
	PageSize   int         `json:"page_size"`   // 每页数量
	TotalPages int         `json:"total_pages"` // 总页数
}

// ================================
// 响应方法
// ================================

// Success 成功响应
// ErrorCode 为 0 表示成功
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		ErrorCode: 0,
		Data:      data,
		Message:   "ok",
	})
}

// Fail 失败响应
// ErrorCode 不为 0 表示失败
func Fail(c *gin.Context, errorCode int, msg string) {
	c.JSON(http.StatusOK, Response{
		ErrorCode: errorCode,
		Data:      nil,
		Message:   msg,
	})
}

// FailByError 失败响应（使用自定义错误）
func FailByError(c *gin.Context, err global.CustomError) {
	Fail(c, err.ErrorCode, err.ErrorMsg)
}

// ValidateFail 参数验证失败响应
func ValidateFail(c *gin.Context, msg string) {
	Fail(c, global.Errors.ValidateError.ErrorCode, msg)
}

// BusinessFail 业务逻辑失败响应
func BusinessFail(c *gin.Context, msg string) {
	Fail(c, global.Errors.BusinessError.ErrorCode, msg)
}

// TokenFail Token 验证失败响应
func TokenFail(c *gin.Context) {
	FailByError(c, global.Errors.TokenError)
}

// ServerError 服务器内部错误响应
func ServerError(c *gin.Context, err interface{}) {
	msg := "Internal Server Error"
	// 非生产环境显示具体错误信息
	if os.Getenv(gin.EnvGinMode) != gin.ReleaseMode {
		if e, ok := err.(error); ok {
			msg = e.Error()
		}
	}
	c.JSON(http.StatusInternalServerError, Response{
		ErrorCode: http.StatusInternalServerError,
		Data:      nil,
		Message:   msg,
	})
	c.Abort()
}

// ServerErrorWithConfig 服务器内部错误响应（使用配置）
func ServerErrorWithConfig(c *gin.Context, err interface{}, cfg *config.Configuration) {
	msg := "Internal Server Error"
	// 非生产环境显示具体错误信息
	if cfg.App.Env != "production" && os.Getenv(gin.EnvGinMode) != gin.ReleaseMode {
		if e, ok := err.(error); ok {
			msg = e.Error()
		}
	}
	c.JSON(http.StatusInternalServerError, Response{
		ErrorCode: http.StatusInternalServerError,
		Data:      nil,
		Message:   msg,
	})
	c.Abort()
}
