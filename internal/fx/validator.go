package fx

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"gin-web/utils"
)

// initValidator 注册自定义校验器（mobile/email）。
//
// fx 重构后未再调用 bootstrap.InitializeValidator，导致带 mobile/email 等 tag 的
// 请求在 ShouldBindJSON 时因校验函数未注册而 panic（"Undefined validation function"）。
// 这里在 Gin 引擎初始化时补上注册。
//
// 注意：不注册 RegisterTagNameFunc。dto.GetErrorMsg 用 v.Field()+v.Tag() 查找自定义
// 文案，而 GetMessages 的 key 用的是 Go 字段名（如 "Mobile.mobile"）；若注册了
// RegisterTagNameFunc，v.Field() 会变成 json tag 名（"mobile"），导致查不到自定义文案，
// 回退成英文原文。保持默认（Go 字段名）才能与 GetMessages 对齐。
func initValidator() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	_ = v.RegisterValidation("mobile", utils.ValidateMobile)
	_ = v.RegisterValidation("email", utils.ValidateEmail)
}

