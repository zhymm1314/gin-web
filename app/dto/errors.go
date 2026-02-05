package dto

// ================================
// 统一错误码定义
// ================================

// 错误码常量
const (
	// 成功
	CodeSuccess = 0

	// 业务错误 4xxxx
	CodeBusinessError = 40000
	CodeTokenError    = 40100
	CodeValidateError = 42200

	// 服务器错误 5xxxx
	CodeServerError = 50000
)

// CustomError 自定义错误结构
type CustomError struct {
	ErrorCode int
	ErrorMsg  string
}

// 预定义错误
var (
	ErrBusiness = CustomError{CodeBusinessError, "业务错误"}
	ErrValidate = CustomError{CodeValidateError, "请求参数错误"}
	ErrToken    = CustomError{CodeTokenError, "登录授权失效"}
)
