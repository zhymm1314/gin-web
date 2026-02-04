package controllers

import (
	"gin-web/app/common/request"
	"gin-web/app/common/response"
	"gin-web/app/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ModController mod控制器
type ModController struct {
	modService *services.ModService
}

// NewModController 创建Mod控制器实例
func NewModController(modService *services.ModService) *ModController {
	return &ModController{modService: modService}
}

// Prefix 返回路由前缀
func (mc *ModController) Prefix() string {
	return ""
}

// Routes 返回路由列表
func (mc *ModController) Routes() []Route {
	return []Route{
		{Method: "GET", Path: "/mods/search", Handler: mc.Search},
		{Method: "GET", Path: "/mods/:id", Handler: mc.Detail},
		{Method: "GET", Path: "/mods/:id/download", Handler: mc.Download},
		{Method: "GET", Path: "/games", Handler: mc.Games},
		{Method: "GET", Path: "/categories", Handler: mc.Categories},
	}
}

// Search 搜索mod
func (mc *ModController) Search(c *gin.Context) {
	var req request.ModSearchRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidateFail(c, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	result, err := mc.modService.SearchMods(req)
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Detail 获取mod详情
func (mc *ModController) Detail(c *gin.Context) {
	var req request.ModDetailRequest

	if err := c.ShouldBindUri(&req); err != nil {
		response.ValidateFail(c, err.Error())
		return
	}

	result, err := mc.modService.GetModDetail(req.ID)
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

	result, err := mc.modService.GetModDetail(uint(id))
	if err != nil {
		response.BusinessFail(c, "Mod not found")
		return
	}

	if result.DownloadURL != "" {
		c.Redirect(http.StatusFound, result.DownloadURL)
		return
	}

	response.BusinessFail(c, "Download URL not available")
}

// Games 获取游戏列表
func (mc *ModController) Games(c *gin.Context) {
	result, err := mc.modService.GetGames()
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Categories 获取分类列表
func (mc *ModController) Categories(c *gin.Context) {
	result, err := mc.modService.GetCategories()
	if err != nil {
		response.BusinessFail(c, err.Error())
		return
	}

	response.Success(c, result)
}
