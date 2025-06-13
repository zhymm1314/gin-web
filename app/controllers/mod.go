package app

import (
	"gin-web/app/common/request"
	"gin-web/app/common/response"
	"gin-web/app/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ModController mod控制器
type ModController struct{}

// Search 搜索mod
func (mc *ModController) Search(c *gin.Context) {
	var req request.ModSearchRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidateFail(c, err.Error())
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 调用服务层
	result, err := services.ModService.SearchMods(req)
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Detail 获取mod详情
func (mc *ModController) Detail(c *gin.Context) {
	var req request.ModDetailRequest

	// 绑定URI参数
	if err := c.ShouldBindUri(&req); err != nil {
		response.ValidateFail(c, err.Error())
		return
	}

	// 调用服务层
	result, err := services.ModService.GetModDetail(req.ID)
	if err != nil {
		response.BusinessFail(c, "Mod not found")
		return
	}

	response.Success(c, result)
}

// Download 下载mod
func (mc *ModController) Download(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ValidateFail(c, "Invalid mod ID")
		return
	}

	// 获取mod详情
	result, err := services.ModService.GetModDetail(uint(id))
	if err != nil {
		response.BusinessFail(c, "Mod not found")
		return
	}

	// 如果有下载链接，重定向到下载地址
	if result.DownloadURL != "" {
		c.Redirect(http.StatusFound, result.DownloadURL)
		return
	}

	response.BusinessFail(c, "Download URL not available")
}

// Games 获取游戏列表
func (mc *ModController) Games(c *gin.Context) {
	result, err := services.ModService.GetGames()
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Categories 获取分类列表
func (mc *ModController) Categories(c *gin.Context) {
	result, err := services.ModService.GetCategories()
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}

	response.Success(c, result)
}
