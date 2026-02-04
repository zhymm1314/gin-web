package controllers

import (
	"gin-web/app/common/request"
	"gin-web/app/common/response"
	"gin-web/app/services"
	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var form request.Register
	if err := c.ShouldBindJSON(&form); err != nil {
		response.ValidateFail(c, request.GetErrorMsg(form, err))
		return
	}

	user, err := services.UserServiceLegacy.Register(form)
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}
	response.Success(c, user)
}
