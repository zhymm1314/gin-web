package utils

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

// ValidateMobile 校验手机号
func ValidateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	ok, _ := regexp.MatchString(`^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\d{8}$`, mobile)
	if !ok {
		return false
	}
	return true
}

// ValidateE-mail 校验邮箱
func ValidateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	ok, _ := regexp.MatchString(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`, email)
	if !ok {
		return false
	}
	return true
}
