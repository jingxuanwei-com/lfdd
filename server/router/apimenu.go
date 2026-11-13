package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apimenu 菜单管理接口：新建/改名/删除/顺序整组重写，均须登录（AuthGuard）
func apiMenuRouter() {
	apiMenuHandler := &handler.ApiMenuHandler{}
	group := web.Engine.Group("/api/menu")
	group.POST("/create", middleware.AuthGuard(), apiMenuHandler.Create)
	group.POST("/update", middleware.AuthGuard(), apiMenuHandler.Update)
	group.POST("/delete", middleware.AuthGuard(), apiMenuHandler.Delete)
	group.POST("/sort", middleware.AuthGuard(), apiMenuHandler.Sort)
}
