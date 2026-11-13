package handler

import (
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApiGoodsHandler struct {
}

// Sidebar 商品侧边菜单（菜单/分类列表，按 sort 升序）
func (h *ApiGoodsHandler) Sidebar(c *gin.Context) {
	out, err := service.GetGoodsSidebar(service.GetGoodsSidebarInput{})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// Page 商品页面（指定菜单下上架商品的分页列表）
// query: menu_id（可选，0/不传 = 全部）、page（默认 1）、page_size（默认 20，上限 100）
func (h *ApiGoodsHandler) Page(c *gin.Context) {
	var req struct {
		MenuID   int64 `form:"menu_id"`
		Page     int   `form:"page"`
		PageSize int   `form:"page_size"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.GetGoodsPage(service.GetGoodsPageInput{
		MenuID:   req.MenuID,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}
