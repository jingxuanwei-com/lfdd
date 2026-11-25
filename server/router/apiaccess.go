package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apiAccess 访问控制接口：读取/保存开关、增删令牌仅管理员（RoleGuard）；
// verify 为公开接口，前端路由守卫凭它判断当前浏览器是否有访问资格。
func apiAccessRouter() {
	apiAccessHandler := &handler.ApiAccessHandler{}
	group := web.Engine.Group("/api/access")

	group.GET("/settings", middleware.RoleGuard(), apiAccessHandler.Settings)
	group.POST("/save", middleware.RoleGuard(), apiAccessHandler.Save)
	group.POST("/token/create", middleware.RoleGuard(), apiAccessHandler.TokenCreate)
	group.POST("/token/delete", middleware.RoleGuard(), apiAccessHandler.TokenDelete)
	group.POST("/token/clear-expired", middleware.RoleGuard(), apiAccessHandler.TokenClearExpired)
	group.POST("/verify", apiAccessHandler.Verify)
}
