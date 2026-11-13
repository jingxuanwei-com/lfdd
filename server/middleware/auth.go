package middleware

import (
	"errors"
	"strings"

	"server/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============ 登录拦截系列中间件（AuthGuard / GuestGuard / OptionalGuard） ============

// tokenFromHeader 解析 Authorization: Bearer <token>（只认 Authorization 头，避免读 body 破坏后续参数绑定）
func tokenFromHeader(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
}

// lookupSession 校验 token 对应的会话，返回所属用户 ID。
// ok=false 表示未登录（token 为空、会话不存在或已过期）；err 为数据库等系统异常。
func lookupSession(token string) (userID int64, ok bool, err error) {
	if token == "" {
		return 0, false, nil
	}
	s, err := service.GetSession(service.GetSessionInput{Token: token})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return s.Session.UserID, true, nil
}

// AuthGuard 登录拦截中间件：**登录放行，未登录阻止**。
// 解析 Authorization: Bearer <token> → 校验会话 → 通过后把 userID、token 写入上下文
// （handler 用 c.GetInt64("userID") / c.GetString("token") 取）；
// 未登录/会话失效返回 401 并中断请求，数据库异常返回 500。
// 用于必须登录才能访问的接口。
func AuthGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromHeader(c)
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"message": "未登录"})
			return
		}

		userID, ok, err := lookupSession(token)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"message": "服务器内部错误"})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"message": "登录已过期或会话无效"})
			return
		}

		// 登录放行：把 userID、token 写入上下文供 handler 使用
		c.Set("userID", userID)
		c.Set("token", token)
		c.Next()
	}
}

// GuestGuard 未登录拦截中间件（游客中间件）：**未登录放行，已登录阻止**。
// 用于登录、注册等已登录用户反而不应访问的页面：
// - 未携带 token，或会话不存在/已过期 → 视为未登录，放行；
// - 会话有效（已登录）→ 403 中断请求。
func GuestGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok, err := lookupSession(tokenFromHeader(c))
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"message": "服务器内部错误"})
			return
		}
		if !ok {
			// 未登录（含 token 失效）→ 放行
			c.Next()
			return
		}

		// 已登录 → 阻止
		c.AbortWithStatusJSON(403, gin.H{"message": "已登录，禁止访问该页面"})
	}
}

// OptionalGuard 可选登录中间件：**不管是否登录都放行**，登录了就顺带把身份带入上下文。
// - 已登录 → 把 userID、token 写入上下文（c.GetInt64("userID") / c.GetString("token")）后放行；
// - 未登录（无 token 或会话失效）→ 直接放行，不写上下文；
// handler 用 c.GetInt64("userID") != 0 判断是否已登录。
// 用于登录与否都能访问、但登录后需要带出用户信息（个性化）的接口。
// 会话查询异常也不拦截：可选登录不应因查询失败挡住公开请求。
func OptionalGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromHeader(c)
		userID, ok, err := lookupSession(token)
		if err != nil {
			c.Next()
			return
		}
		if ok {
			// 已登录：带入身份
			c.Set("userID", userID)
			c.Set("token", token)
		}
		c.Next()
	}
}
