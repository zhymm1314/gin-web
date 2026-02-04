package controllers

import (
	"gin-web/app/common/request"
	"gin-web/app/common/response"
	"gin-web/app/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Login(c *gin.Context) {
	var form request.Login
	if err := c.ShouldBindJSON(&form); err != nil {
		response.ValidateFail(c, request.GetErrorMsg(form, err))
		return
	}

	user, err := services.UserServiceLegacy.Login(form)
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}
	tokenData, _, err := services.JwtService.CreateToken(services.AppGuardName, user)
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}
	response.Success(c, tokenData)
}

func Info(c *gin.Context) {
	user, err := services.UserServiceLegacy.GetUserInfo(c.Keys["id"].(string))
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}
	response.Success(c, user)
}

func Logout(c *gin.Context) {
	err := services.JwtService.JoinBlackList(c.Keys["token"].(*jwt.Token))
	if err != nil {
		response.BusinessFail(c, "登出失败")
		return
	}
	response.Success(c, nil)
}
