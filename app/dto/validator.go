package dto

import (
	"github.com/go-playground/validator/v10"
)

// Validator 验证器接口
// 实现此接口可自定义字段验证错误信息
type Validator interface {
	GetMessages() ValidatorMessages
}

// ValidatorMessages 验证消息映射
// key 格式: "字段名.规则名"，如 "Mobile.required"
type ValidatorMessages map[string]string

// GetErrorMsg 获取验证错误信息
// 如果 request 实现了 Validator 接口，则使用自定义错误信息
func GetErrorMsg(request interface{}, err error) string {
	if _, isValidatorErrors := err.(validator.ValidationErrors); isValidatorErrors {
		_, isValidator := request.(Validator)
		for _, v := range err.(validator.ValidationErrors) {
			// 若 request 结构体实现 Validator 接口即可实现自定义错误信息
			if isValidator {
				if message, exist := request.(Validator).GetMessages()[v.Field()+"."+v.Tag()]; exist {
					return message
				}
			}
			return v.Error()
		}
	} else {
		return err.Error()
	}

	return "Parameter error"
}
