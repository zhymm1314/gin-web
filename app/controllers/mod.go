package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gin-web/app/dto"
	"gin-web/app/services"
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
// @Summary      搜索 Mod
// @Description  根据关键词、游戏、分类等条件搜索 Mod
// @Tags         Mod
// @Accept       json
// @Produce      json
// @Param        keyword query string false "搜索关键词"
// @Param        game_id query int false "游戏ID"
// @Param        category_id query int false "分类ID"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200 {object} dto.Response "成功"
// @Failure      400 {object} dto.Response "参数错误"
// @Router       /mods/search [get]
func (mc *ModController) Search(c *gin.Context) {
	var req dto.ModSearchRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ValidateFail(c, err.Error())
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
		dto.BusinessFail(c, err.Error())
		return
	}

	dto.Success(c, result)
}

// Detail 获取mod详情
// @Summary      获取 Mod 详情
// @Description  根据 ID 获取 Mod 详细信息
// @Tags         Mod
// @Accept       json
// @Produce      json
// @Param        id path int true "Mod ID"
// @Success      200 {object} dto.Response "成功"
// @Failure      404 {object} dto.Response "未找到"
// @Router       /mods/{id} [get]
func (mc *ModController) Detail(c *gin.Context) {
	var req dto.ModDetailRequest

	if err := c.ShouldBindUri(&req); err != nil {
		dto.ValidateFail(c, err.Error())
		return
	}

	result, err := mc.modService.GetModDetail(req.ID)
	if err != nil {
		dto.BusinessFail(c, "Mod not found")
		return
	}

	dto.Success(c, result)
}

// Download 下载mod
// @Summary      下载 Mod
// @Description  获取 Mod 下载链接并重定向
// @Tags         Mod
// @Param        id path int true "Mod ID"
// @Success      302 {string} string "重定向到下载链接"
// @Failure      404 {object} dto.Response "未找到"
// @Router       /mods/{id}/download [get]
func (mc *ModController) Download(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		dto.ValidateFail(c, "Invalid mod ID")
		return
	}

	result, err := mc.modService.GetModDetail(uint(id))
	if err != nil {
		dto.BusinessFail(c, "Mod not found")
		return
	}

	if result.DownloadURL != "" {
		c.Redirect(http.StatusFound, result.DownloadURL)
		return
	}

	dto.BusinessFail(c, "Download URL not available")
}

// Games 获取游戏列表
// @Summary      获取游戏列表
// @Description  获取所有支持的游戏列表
// @Tags         Mod
// @Produce      json
// @Success      200 {object} dto.Response "成功"
// @Router       /games [get]
func (mc *ModController) Games(c *gin.Context) {
	result, err := mc.modService.GetGames()
	if err != nil {
		dto.BusinessFail(c, err.Error())
		return
	}

	dto.Success(c, result)
}

// Categories 获取分类列表
// @Summary      获取分类列表
// @Description  获取所有 Mod 分类列表
// @Tags         Mod
// @Produce      json
// @Success      200 {object} dto.Response "成功"
// @Router       /categories [get]
func (mc *ModController) Categories(c *gin.Context) {
	result, err := mc.modService.GetCategories()
	if err != nil {
		dto.BusinessFail(c, err.Error())
		return
	}

	dto.Success(c, result)
}
