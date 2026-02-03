package errors

import "fmt"

// BizError 业务错误
type BizError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *BizError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *BizError) Unwrap() error {
	return e.Err
}

// New 创建业务错误
func New(code int, message string) *BizError {
	return &BizError{Code: code, Message: message}
}

// Wrap 包装错误
func Wrap(err error, code int, message string) *BizError {
	return &BizError{Code: code, Message: message, Err: err}
}

// 预定义错误码
const (
	CodeSuccess         = 0
	CodeValidationError = 42200
	CodeUnauthorized    = 40100
	CodeForbidden       = 40300
	CodeNotFound        = 40400
	CodeBusinessError   = 40000
	CodeInternalError   = 50000

	// 用户相关
	CodeUserNotFound  = 20001
	CodeUserExists    = 20002
	CodePasswordError = 20003
)

// 预定义错误
var (
	ErrValidation   = New(CodeValidationError, "参数验证失败")
	ErrUnauthorized = New(CodeUnauthorized, "登录授权失效")
	ErrForbidden    = New(CodeForbidden, "禁止访问")
	ErrNotFound     = New(CodeNotFound, "资源不存在")
	ErrInternal     = New(CodeInternalError, "服务器内部错误")
	ErrBusiness     = New(CodeBusinessError, "业务错误")
	ErrUserNotFound = New(CodeUserNotFound, "用户不存在")
	ErrUserExists   = New(CodeUserExists, "用户已存在")
	ErrPassword     = New(CodePasswordError, "密码错误")
)
