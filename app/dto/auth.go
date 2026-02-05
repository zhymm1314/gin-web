package dto

// ================================
// 认证模块 DTO（Data Transfer Object）
// ================================

// -------------------- Request --------------------

// RegisterRequest 用户注册请求
// @Description 用户注册所需信息
type RegisterRequest struct {
	Name     string `form:"name" json:"name" binding:"required" example:"张三"`                       // 用户名称
	Mobile   string `form:"mobile" json:"mobile" binding:"required,mobile" example:"13800138000"`   // 手机号码
	Password string `form:"password" json:"password" binding:"required" example:"123456"`           // 登录密码
	Email    string `form:"email" json:"email" binding:"required,email" example:"user@example.com"` // 邮箱地址
}

// GetMessages 自定义验证错误信息
func (r RegisterRequest) GetMessages() ValidatorMessages {
	return ValidatorMessages{
		"Name.required":     "用户名称不能为空",
		"Mobile.required":   "手机号码不能为空",
		"Mobile.mobile":     "手机号码格式不正确",
		"Password.required": "用户密码不能为空",
		"Email.required":    "邮箱不能为空",
		"Email.email":       "邮箱格式不正确",
	}
}

// LoginRequest 用户登录请求
// @Description 用户登录凭证
type LoginRequest struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required,mobile" example:"13800138000"` // 手机号码
	Password string `form:"password" json:"password" binding:"required" example:"123456"`         // 登录密码
}

// GetMessages 自定义验证错误信息
func (r LoginRequest) GetMessages() ValidatorMessages {
	return ValidatorMessages{
		"Mobile.required":   "手机号码不能为空",
		"Mobile.mobile":     "手机号码格式不正确",
		"Password.required": "用户密码不能为空",
	}
}

// -------------------- Response --------------------

// LoginResponse 登录成功响应
// @Description 登录成功返回的 Token 信息
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."` // JWT Token
	ExpiresIn   int64  `json:"expires_in" example:"43200"`                     // 过期时间（秒）
	TokenType   string `json:"token_type" example:"Bearer"`                    // Token 类型
}

// UserInfoResponse 用户信息响应
// @Description 当前登录用户的详细信息
type UserInfoResponse struct {
	ID        uint   `json:"id" example:"1"`                           // 用户ID
	Name      string `json:"name" example:"张三"`                        // 用户名称
	Mobile    string `json:"mobile" example:"138****8000"`             // 手机号码（脱敏）
	Email     string `json:"email" example:"user@example.com"`         // 邮箱地址
	CreatedAt string `json:"created_at" example:"2024-01-01 12:00:00"` // 注册时间
}
