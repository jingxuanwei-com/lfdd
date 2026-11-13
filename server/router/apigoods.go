package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apigoods 商品接口：侧边菜单 + 商品页面，均免登录访问
// （OptionalGuard：未登录放行、登录顺带把 userID/token 写入上下文）
func apiGoodsRouter() {
	apiGoodsHandler := &handler.ApiGoodsHandler{}
	group := web.Engine.Group("/api/goods")
	group.GET("/sidebar", middleware.OptionalGuard(), apiGoodsHandler.Sidebar)
	group.GET("/page", middleware.OptionalGuard(), apiGoodsHandler.Page)
}
