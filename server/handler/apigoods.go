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

// Create 新建商品（挂到指定菜单下）
// body: {"menu_id":"101","name":"杨枝甘露","price":1850,"spec":"","stock":-1,"image":"","status":1,"sort":0}
// price 单位：分（1850 = 18.50 元）
func (h *ApiGoodsHandler) Create(c *gin.Context) {
	var req service.CreateGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.CreateGoods(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// Update 修改商品（价格、顺序、库存、上下架、所属菜单等全量更新，前端回传完整对象）
func (h *ApiGoodsHandler) Update(c *gin.Context) {
	var req service.UpdateGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.UpdateGoods(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Delete 删除商品
func (h *ApiGoodsHandler) Delete(c *gin.Context) {
	var req service.DeleteGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.DeleteGoods(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Sort 商品顺序整组重写（前端拖拽后调用）
// body: {"ids":["1","2","3"]} 按数组下标写 sort（0,1,2...）
func (h *ApiGoodsHandler) Sort(c *gin.Context) {
	var req service.SortGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.SortGoods(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}
