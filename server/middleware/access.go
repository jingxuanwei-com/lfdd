package middleware

import (
	"errors"
	"net/http"
	"strings"

	"server/model"
	"server/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// accessExemptPaths 访问令牌中间件的豁免路径：
// 登录（/admin 配置入口依赖）、访问校验（前端路由守卫使用）、探活接口。
var accessExemptPaths = map[string]bool{
	"/api/auth/login":    true,
	"/api/access/verify": true,
	"/api/":              true,
}

// AccessGuard 访问令牌中间件：开启访问限制后，未授权请求返回伪装的纯文本 404，
// 与前端未授权页保持一致；放行规则复用 service.VerifyAccessToken
// （允许直接访问 / 有效登录会话 / 有效访问令牌 X-Access-Token）。
// 图片直出（/api/upload/*）豁免：<img> 请求无法携带认证头。
// 已携带失效登录会话的请求仍返回 401，前端据此清理登录态，避免体验变化。
func AccessGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if accessExemptPaths[path] || strings.HasPrefix(path, "/api/upload/") {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		out, err := service.VerifyAccessToken(service.VerifyAccessTokenInput{
			Token:         c.GetHeader("X-Access-Token"),
			Authorization: auth,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "服务器内部错误",
			})
			return
		}
		if out.Allowed {
			c.Next()
			return
		}

		// 失效的登录会话：保持 401 语义，前端据此退出登录
		token := bearerTokenOf(auth)
		if token != "" && !strings.HasPrefix(token, model.GuestTokenPrefix) {
			if _, err = service.GetSession(service.GetSessionInput{Token: token}); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"message": "登录已过期或会话无效",
					})
					return
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "服务器内部错误",
				})
				return
			}
		}

		c.Abort()
		c.Data(http.StatusNotFound, "text/plain; charset=utf-8", []byte("404 page not found"))
	}
}

// bearerTokenOf 解析 Authorization: Bearer <token>，没有则返回空串
func bearerTokenOf(h string) string {
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}
