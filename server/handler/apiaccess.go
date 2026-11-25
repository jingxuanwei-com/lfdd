package handler

import (
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApiAccessHandler struct {
}

// Settings 读取访问控制配置与令牌列表
func (h *ApiAccessHandler) Settings(c *gin.Context) {
	out, err := service.GetAccessSettings(service.GetAccessSettingsInput{})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// Save 保存「是否允许直接访问」与网站地址（body: {"allow_direct":true,"site_url":"https://example.com"}）
func (h *ApiAccessHandler) Save(c *gin.Context) {
	var req service.SaveAccessSettingsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.SaveAccessSettings(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// TokenCreate 生成访问令牌（body: {"expire_hours":24}，0 表示永久）
func (h *ApiAccessHandler) TokenCreate(c *gin.Context) {
	var req service.CreateAccessTokenInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.CreateAccessToken(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// TokenDelete 删除访问令牌（body: {"id":"1"}）
func (h *ApiAccessHandler) TokenDelete(c *gin.Context) {
	var req service.DeleteAccessTokenInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.DeleteAccessToken(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Verify 校验访问令牌（公开接口，前端路由守卫调用；body: {"token":"..."}）。
// 携带有效登录会话（Authorization 头）的用户免令牌放行；游客身份不算登录。
func (h *ApiAccessHandler) Verify(c *gin.Context) {
	var req service.VerifyAccessTokenInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}
	req.Authorization = c.GetHeader("Authorization")

	out, err := service.VerifyAccessToken(req)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}
